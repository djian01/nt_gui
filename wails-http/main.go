package main

import (
	"embed"
	"log"
	"sync"

	"github.com/djian01/nt_gui/wails-http/internal/ping"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

// PingService is the desktop bridge. The runner has no Wails dependency.
type PingService struct {
	runner  *ping.Runner
	app     *application.App
	windows sync.Mutex
}

func (s *PingService) Start(config ping.Config) (ping.Session, error) { return s.runner.Start(config) }
func (s *PingService) Restart(id string) (ping.Session, error)        { return s.runner.Restart(id) }
func (s *PingService) List() []ping.Session                           { return s.runner.List() }
func (s *PingService) Get(id string) (ping.Detail, error)             { return s.runner.Get(id) }
func (s *PingService) Stop(id string) (ping.Session, error)           { return s.runner.Stop(id) }
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
	service.runner = ping.New(func(s ping.Session) { service.app.Event.Emit("http:updated", s) })
	app := application.New(application.Options{
		Name: "NT HTTP Prototype", Description: "Cross-platform HTTP diagnostics",
		Services:   []application.Service{application.NewService(service)},
		Assets:     application.AssetOptions{Handler: application.BundledAssetFileServer(assets)},
		OnShutdown: service.runner.Close,
		Mac:        application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	service.app = app
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
