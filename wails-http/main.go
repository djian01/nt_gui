package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/djian01/nt_gui/wails-http/internal/ping"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

// PingService is the desktop bridge. The runner has no Wails dependency.
type PingService struct {
	runner  *ping.Runner
	store   *ping.Store
	app     *application.App
	windows sync.Mutex
}

func (s *PingService) Start(config ping.Config) (ping.Session, error) { return s.runner.Start(config) }
func (s *PingService) Restart(id, password string) (ping.Session, error) {
	return s.runner.Restart(id, password)
}
func (s *PingService) List(search, filter string, before int64) (ping.Page, error) {
	return s.runner.List(search, filter, before)
}
func (s *PingService) Get(id string) (ping.Detail, error)   { return s.runner.Get(id) }
func (s *PingService) Stop(id string) (ping.Session, error) { return s.runner.Stop(id) }
func (s *PingService) Remove(id string) error {
	s.windows.Lock()
	defer s.windows.Unlock()
	if err := s.runner.Remove(id); err != nil {
		return err
	}
	if w, ok := s.app.Window.GetByName("chart-" + id); ok {
		w.Close()
	}
	s.app.Event.Emit("http:removed", id)
	return nil
}

func (s *PingService) Timeline(id string, from, to int64) (ping.Timeline, error) {
	return s.store.Timeline(id, from, to)
}

func (s *PingService) ExportCSV(id string) (string, error) {
	if _, err := s.store.Get(id); err != nil {
		return "", err
	}
	path, err := s.app.Dialog.SaveFile().SetFilename("HTTP-"+time.Now().Format("20060102-150405")+".csv").AddFilter("CSV results", "*.csv").SetMessage("Export all saved probes").PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	// Prepare a complete export before replacing a user-selected destination.
	file, err := os.CreateTemp(filepath.Dir(path), ".nt-export-*.csv")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err = s.store.ExportCSV(id, file); err != nil {
		return "", err
	}
	if err = file.Sync(); err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *PingService) OpenChart(id string) error {
	s.windows.Lock()
	defer s.windows.Unlock()
	if _, err := s.runner.Get(id); err != nil {
		return err
	}
	if w, ok := s.app.Window.GetByName("chart-" + id); ok {
		w.Show()
		w.Focus()
		return nil
	}
	s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "chart-" + id, Title: "HTTP Ping · Latency", URL: "/?chart=" + id,
		Width: 1100, Height: 760, MinWidth: 780, MinHeight: 600,
		BackgroundColour: application.NewRGB(11, 18, 32),
		Mac:              application.MacWindow{Appearance: application.NSAppearanceNameDarkAqua},
		Windows:          application.WindowsWindow{Theme: application.Dark},
	})
	return nil
}

func main() {
	service := &PingService{}
	application.RegisterEvent[ping.Session]("http:updated")
	application.RegisterEvent[string]("http:removed")
	app := application.New(application.Options{
		Name: "NT HTTP Prototype", Description: "Cross-platform HTTP diagnostics",
		Services: []application.Service{application.NewService(service)},
		Assets:   application.AssetOptions{Handler: application.BundledAssetFileServer(assets)},
		OnShutdown: func() {
			if service.runner != nil {
				if err := service.runner.Close(); err != nil {
					log.Printf("Save on shutdown: %v", err)
				}
			}
			if service.store != nil {
				if err := service.store.Close(); err != nil {
					log.Printf("Close results: %v", err)
				}
			}
		},
		SingleInstance: &application.SingleInstanceOptions{UniqueID: "net.packetstreams.ntgui.wails", OnSecondInstanceLaunch: func(application.SecondInstanceData) {
			if service.app != nil {
				if w, ok := service.app.Window.GetByName("main"); ok {
					w.Show()
					w.Focus()
				}
			}
		}},
		Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	service.app = app
	path, err := ping.DefaultStorePath()
	if dir := os.Getenv("NT_WAILS_DATA_DIR"); dir != "" {
		path = filepath.Join(dir, "results.db")
		err = nil
	}
	if err != nil {
		log.Fatal(err)
	}
	service.store, err = ping.OpenStore(path)
	if err != nil {
		log.Fatalf("Cannot open saved results: %v", err)
	}
	service.runner = ping.New(service.store, func(s ping.Session) { app.Event.Emit("http:updated", s) })
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "main", Title: "NT · HTTP Ping", URL: "/", Width: 1420, Height: 960,
		MinWidth: 980, MinHeight: 720, BackgroundColour: application.NewRGB(11, 18, 32),
		Mac:     application.MacWindow{Appearance: application.NSAppearanceNameDarkAqua},
		Windows: application.WindowsWindow{Theme: application.Dark},
	})
	window.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) { app.Quit() })
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
