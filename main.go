package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

// GUI mode (no terminal attached) for Windows ONLY
// go build -ldflags="-H=windowsgui" -o nt_gui.exe

// create a global logger pointer
var (
	logger *log.Logger
)

// app Version
var appVersion string = "1.0.2"

// Create a global cancelable context
var appCtx, cancelFunc = context.WithCancel(context.Background())

// test register: records all the active testing UUIDs
var testRegister []string

func main() {

	// get the config file path
	// macOS: ~/Library/Application Support/<appName>
	// Windows & Linux: the config file path is the same as the executable path
	configPath, err := getConfigFilePath("nt_gui")
	if err != nil {
		log.Fatal("Failed to get log file path:", err)
		return
	}

	// create or open the output.txt file for logging
	// "os.O_RDWR": open file to read and write
	// "os.O_CREATE": Create the file with the mode permissions if file does not exist. Cursor is at the beginning.
	// "os.O_APPEND": Only allow write past end of file
	logFile, err := os.OpenFile(filepath.Join(configPath, "logFile.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
		return
	}
	defer logFile.Close()

	// create a new logger
	logger = log.New(logFile, "", log.LstdFlags)

	// create an error chan and error handling
	errChan := make(chan error, 5)
	defer close(errChan)

	// launch the error catch go routine
	go func() {
		loopClose := false
		for {
			// check loopClose Flag
			if loopClose {
				break
			}

			// select option
			select {

			// ends this test when app is closing
			case <-appCtx.Done():
				loopClose = true

			// harvest the Probe results
			case err := <-errChan:
				logger.Println(err)
			default:
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()

	//// defer func() to capture the panic & debug stack messages
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered panic: %v", r)
			stack := debug.Stack()
			logger.Printf("Stack Trace: %v", string(stack))
		}
	}()

	a := app.NewWithID("net.packetstreams.ntgui")

	// Ensure cleanup when the app closes
	a.Lifecycle().SetOnStopped(func() {
		logger.Println("Application stopping gracefully.")
		// call cancelFunc() to signal the go routines
		cancelFunc()
		// give time for go routines to exit
		time.Sleep(1 * time.Second)
	})

	// Capture Ctrl+C only for Linux/Mac, skip for Windows
	if runtime.GOOS != "windows" {
		signalChan := make(chan os.Signal, 1)
		defer close(signalChan)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			for s := range signalChan {
				logger.Println("Received shutdown signal (Ctrl+C),", s)
				a.Quit() // Trigger app stop
			}
		}()
	}

	w := a.NewWindow("NT GUI") // w is a pointer

	// Intercept close to ensure app exits, the app will stop once the main window is closed
	w.SetCloseIntercept(func() {
		a.Quit()
	})

	// set window size, default: max window
	screenSize, err := getPrimaryScreenSize()
	if err != nil {
		//log.Println("Failed to get screen size:", err)
		screenSize = fyne.NewSize(1650, 950) // fallback
	}

	w.Resize(screenSize)
	w.CenterOnScreen()

	// Open NT DB
	ntDB, err := ntdb.DBOpen(filepath.Join(configPath, "ntdata.db"))
	if err != nil {
		logger.Println(err)
	}
	defer ntDB.Close()

	// Entry Chan with buffer 10
	entryChan := make(chan ntdb.DbEntry, 10)
	defer close(entryChan)

	// run Insert Entry Go routine
	go ntdb.InsertEntry(ntDB, entryChan, errChan)

	// make UI
	makeUI(w, a, ntDB, entryChan, errChan)

	w.ShowAndRun()

}
