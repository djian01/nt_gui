package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/djian01/nt/pkg/ntPinger"
)

func ResultAnalysisContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// Initial slides
	inputResultPackets := []ntPinger.Packet{}

	// Input Result CSV File card
	inputResultCSVFilePath := widget.NewEntry()
	inputResultCSVFilePath.SetPlaceHolder("please input the file path or press the right button to select the Result CSV file")
	inputResultCSVFileButton := widget.NewButton("Select the Result CSV File", func() {})
	inputResultCSVFileButton.Importance = widget.WarningImportance
	inputNSXConfigContainer := container.New(layout.NewBorderLayout(nil, nil, nil, inputResultCSVFileButton), inputResultCSVFilePath, inputResultCSVFileButton)
	inputResultCSVFileCard := widget.NewCard("", "Input the existing Result CSV File", inputNSXConfigContainer)

	// Input NSX Config File BTN
	inputResultCSVFileButton.OnTapped = OpenResultCSVFile(w, &inputResultPackets)

	//// Main Container
	spaceHolder := widget.NewLabel("                     ")
	RaMainContainerInner := container.New(layout.NewVBoxLayout(), inputResultCSVFileCard)
	RaMainContainerOuter := container.New(layout.NewBorderLayout(spaceHolder, spaceHolder, spaceHolder, spaceHolder), spaceHolder, RaMainContainerInner)

	// Return your result analysis interface components here
	return RaMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: OpenResultCSVFile
func OpenResultCSVFile(w fyne.Window, inputResultPackets *[]ntPinger.Packet) func() {
	return func() {

		//// Select Analysis Dialog
		RA_Dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {

			// if err when saving, return
			if err != nil {
				logger.Println("Error opening file")
				return
			}

			// if user cancel, return
			if reader == nil {
				// user cancelled
				logger.Println("No file selected")
				return
			}

			defer reader.Close()

			// Open file using Fyne
			r := csv.NewReader(reader)
			records, err := r.ReadAll()
			if err != nil {
				logger.Println("Error reading CSV")
				return
			}

			// Get Type
			RaType := records[1][0]

			fmt.Println(RaType)

		}, w)

		// resize the dialog size
		RA_Dialog.Resize(fyne.Size{800, 600})

		// get current executable path
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		exePathURI, _ := storage.ListerForURI(storage.NewFileURI(exeDir))
		RA_Dialog.SetLocation(exePathURI)

		// create a file extension filter
		filter1 := storage.NewExtensionFileFilter([]string{".csv"})
		RA_Dialog.SetFilter(filter1)

		// show dialog
		RA_Dialog.Show()
	}
}
