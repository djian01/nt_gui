package main

import (
	"log"
	"os"
	"runtime/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// create a global logger pointer
var (
	logger *log.Logger
)

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

	w := a.NewWindow("NT GUI") // w is a pointer

	w.Resize(fyne.NewSize(1500, 850))

	// make UI
	makeUI(w, a)

	w.ShowAndRun()

}
