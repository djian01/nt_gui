package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"errors"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/djian01/nt_gui/desktop/internal/testengine"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/icons/net-test.png
var appIcon []byte

// TestService is the desktop bridge. The runner has no Wails dependency.
type TestService struct {
	runner  *testengine.Runner
	store   *testengine.Store
	app     *application.App
	windows sync.Mutex
}

func (s *TestService) Start(config testengine.Config) (testengine.Session, error) {
	return s.runner.Start(config)
}
func (s *TestService) StartDNS(config testengine.Config, resolvers string) ([]testengine.Session, error) {
	return s.runner.StartDNS(config, resolvers)
}
func (s *TestService) StartICMP(config testengine.Config, targets string) ([]testengine.Session, error) {
	return s.runner.StartICMP(config, targets)
}

func (s *TestService) StartTCP(config testengine.Config, targets string) ([]testengine.Session, error) {
	return s.runner.StartTCP(config, targets)
}
func (s *TestService) Record(id string) (testengine.Session, error) {
	return s.runner.Record(id)
}
func (s *TestService) Dismiss(id string) error {
	s.windows.Lock()
	defer s.windows.Unlock()
	if err := s.runner.Dismiss(id); err != nil {
		return err
	}
	if w, ok := s.app.Window.GetByName("chart-" + id); ok {
		w.Close()
	}
	return nil
}
func (s *TestService) Restart(id, password string) (testengine.Session, error) {
	return s.runner.Restart(id, password)
}
func (s *TestService) List(search, filter string, before int64) (testengine.Page, error) {
	return s.runner.List(search, filter, before)
}
func (s *TestService) Get(id string) (testengine.Detail, error) { return s.runner.Get(id) }
func (s *TestService) Stop(id string) (testengine.Session, error) {
	return s.runner.Stop(id)
}
func (s *TestService) Remove(id string) error {
	s.windows.Lock()
	defer s.windows.Unlock()
	if err := s.runner.Remove(id); err != nil {
		return err
	}
	if w, ok := s.app.Window.GetByName("chart-" + id); ok {
		w.Close()
	}
	s.app.Event.Emit("test:removed", id)
	return nil
}

func (s *TestService) Timeline(id string, from, to int64) (testengine.Timeline, error) {
	return s.runner.Timeline(id, from, to)
}

func (s *TestService) ExportCSV(id string) (string, error) {
	if _, err := s.store.Get(id); err != nil {
		return "", err
	}
	path, err := s.app.Dialog.SaveFile().SetFilename("Net-Test-"+time.Now().Format("20060102-150405")+".csv").AddFilter("CSV results", "*.csv").SetMessage("Export all saved probes").PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	return saveExport(path, func(w io.Writer) error { return s.store.ExportCSV(id, w) })
}

func (s *TestService) ImportCSV() (*testengine.Session, error) {
	path, err := s.app.Dialog.OpenFile().AddFilter("HTTP / DNS / TCP / ICMP CSV results", "*.csv").SetMessage("Import HTTP, DNS, TCP, or ICMP results for analysis").PromptForSingleSelection()
	if err != nil || path == "" {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	session, err := s.runner.ImportCSV(file)
	if err != nil {
		return nil, err
	}
	s.app.Event.Emit("test:updated", session)
	return &session, nil
}

func (s *TestService) ExportChart(id, dataURL string) (string, error) {
	if _, err := s.runner.Get(id); err != nil {
		return "", err
	}
	data, err := chartPNG(dataURL)
	if err != nil {
		return "", err
	}
	path, err := s.app.Dialog.SaveFile().SetFilename("Net-Test-Chart-"+time.Now().Format("20060102-150405")+".png").AddFilter("PNG chart", "*.png").SetMessage("Save the visible chart").PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	return saveExport(path, func(w io.Writer) error { _, err := w.Write(data); return err })
}

func chartPNG(dataURL string) ([]byte, error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) || len(dataURL) > 12<<20 {
		return nil, errors.New("Chart must be a PNG image smaller than 9 MB")
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		return nil, errors.New("Invalid chart image")
	}
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || config.Width*config.Height > 12_000_000 {
		return nil, errors.New("Invalid chart image dimensions")
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return nil, errors.New("Invalid PNG chart")
	}
	return data, nil
}

func saveExport(path string, write func(io.Writer) error) (string, error) {
	// Prepare a complete export before replacing a user-selected destination.
	file, err := os.CreateTemp(filepath.Dir(path), ".nt-export-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err = write(file); err != nil {
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

func (s *TestService) OpenChart(id string) error {
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
		Name: "chart-" + id, Title: "Net Test · Latency", URL: "/?chart=" + id,
		Width: 1100, Height: 760, MinWidth: 780, MinHeight: 600,
		BackgroundColour: application.NewRGB(11, 18, 32),
		Mac:              application.MacWindow{Appearance: application.NSAppearanceNameDarkAqua},
		Windows:          application.WindowsWindow{Theme: application.Dark},
	})
	return nil
}

func main() {
	service := &TestService{}
	application.RegisterEvent[testengine.Session]("test:updated")
	application.RegisterEvent[string]("test:removed")
	app := application.New(application.Options{
		Name: "Net Test", Description: "Cross-platform network diagnostics",
		Icon:     appIcon,
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
	path, err := testengine.DefaultStorePath()
	dataDir := os.Getenv("NET_TEST_DATA_DIR")
	if dataDir == "" {
		dataDir = os.Getenv("NT_WAILS_DATA_DIR")
	}
	if dataDir != "" {
		path = filepath.Join(dataDir, "results.db")
		err = nil
	}
	if err != nil {
		log.Fatal(err)
	}
	service.store, err = testengine.OpenStore(path)
	if err != nil {
		log.Fatalf("Cannot open saved results: %v", err)
	}
	service.runner = testengine.New(service.store, func(s testengine.Session) { app.Event.Emit("test:updated", s) })
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "main", Title: "Net Test", URL: "/", Width: 1420, Height: 960,
		MinWidth: 980, MinHeight: 720, BackgroundColour: application.NewRGB(11, 18, 32),
		Mac:     application.MacWindow{Appearance: application.NSAppearanceNameDarkAqua},
		Windows: application.WindowsWindow{Theme: application.Dark},
	})
	window.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) { app.Quit() })
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
