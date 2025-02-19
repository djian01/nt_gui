package main

import (
	"context"
	"log"
	"os"
	"runtime/debug"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

// create a global logger pointer
var (
	logger *log.Logger
)

// Create a global cancelable context
var appCtx, cancelFunc = context.WithCancel(context.Background())

func main() {

	// create or open the output.txt file for logging
	// "os.O_RDWR": open file to read and write
	// "os.O_CREATE": Create the file with the mode permissions if file does not exist. Cursor is at the beginning.
	// "os.O_APPEND": Only allow write past end of file
	logFile, err := os.OpenFile("logFile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
		return
	}
	defer logFile.Close()

	// create a new logger
	logger = log.New(logFile, "", log.LstdFlags)

	//// defer func() to capture the panic & debug stack messages
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered panic: %v", r)
			stack := debug.Stack()
			logger.Printf("Stack Trace: %v", string(stack))
		}
	}()

	a := app.NewWithID("local.ntgui")

	// Ensure cleanup when the app closes
	a.Lifecycle().SetOnStopped(func() {
		// call cancelFunc() to signal the go routines
		cancelFunc()
		// give time for go routines to exit
		time.Sleep(1 * time.Second)
	})

	w := a.NewWindow("NT GUI") // w is a pointer

	w.Resize(fyne.NewSize(1650, 900))

	// Open NT DB
	ntDB, err := ntdb.DBOpen("ntdata.db")
	if err != nil {
		logger.Println(err)
	}
	defer ntDB.Close()

	// Entry Chan
	entryChan := make(chan ntdb.Entry)
	defer close(entryChan)

	// run Insert Entry Go routine
	go ntdb.InsertEntry(ntDB, entryChan)

	// make UI
	makeUI(w, a, ntDB, entryChan)

	w.ShowAndRun()

}
