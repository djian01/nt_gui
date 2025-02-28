package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/djian01/nt/pkg/ntPinger"

	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

func ResultAnalysisContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// Initial slides
	inputResultPackets := []ntPinger.Packet{}
	chartData := []ntchart.ChartPoint{}

	// Input Result CSV File card
	inputResultCSVFilePath := widget.NewEntry()
	inputResultCSVFilePath.SetPlaceHolder("please press the right button to select the Result CSV file")
	inputResultCSVFileButton := widget.NewButton("Select the Result CSV File", func() {})
	inputResultCSVFileButton.Importance = widget.WarningImportance
	inputNSXConfigContainer := container.New(layout.NewBorderLayout(nil, nil, nil, inputResultCSVFileButton), inputResultCSVFilePath, inputResultCSVFileButton)
	inputResultCSVFileCard := widget.NewCard("", "Input the existing Result CSV File", inputNSXConfigContainer)

	// Summary Card
	Summary := Summary{}
	Summary.UI.Initial()
	Summary.UI.CreateCard()

	// Chart Card (Place Holder)
	chart := Chart{}
	chart.Initial()

	// Slider Card
	slider := Slider{}
	slider.Initial(0, 100, 0, 100)
	slider.chartData = &chartData
	slider.CreateCard()
	slider.sliderCard.Hidden = true

	//// Main Container
	RASpaceHolder := widget.NewLabel("                     ")
	RaMainContainerInner := container.New(layout.NewVBoxLayout(), inputResultCSVFileCard, Summary.UI.summaryCard, chart.chartCard, slider.sliderCard)
	RaMainContainerOuter := container.New(layout.NewBorderLayout(RASpaceHolder, RASpaceHolder, RASpaceHolder, RASpaceHolder), RASpaceHolder, RaMainContainerInner)

	// Input NSX Config File BTN
	inputResultCSVFileButton.OnTapped = OpenResultCSVFile(w, &inputResultPackets, &chartData, &chart, &Summary, inputResultCSVFilePath, &slider)

	// Slider Update
	slider.rangeSlider.OnChanged = func() { slider.update() }

	// Return your result analysis interface components here
	return RaMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: OpenResultCSVFile
func OpenResultCSVFile(w fyne.Window, inputResultPackets *[]ntPinger.Packet, chartData *[]ntchart.ChartPoint, chart *Chart, Summary *Summary, inputResultCSVFilePath *widget.Entry, slider *Slider) func() {
	return func() {

		// reset vars
		*inputResultPackets = []ntPinger.Packet{}
		*chartData = []ntchart.ChartPoint{}

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

			// inputResultCSVFilePath set text
			(*inputResultCSVFilePath).SetText(strings.TrimPrefix(reader.URI().String(), "file://"))

			// Open file using Fyne
			r := csv.NewReader(reader)
			records, err := r.ReadAll()
			if err != nil {
				logger.Println("Error reading CSV")
				return
			}

			// Get Result Analysis File Type
			RaType := records[1][0]
			(*Summary).Type = RaType

			appendPacket(inputResultPackets, RaType, &records, chartData, Summary)

			// Create an image Chart
			// verify the image.Bounds(), e.g. image bounds: (0,0)-(1024,512) is good. code -> fmt.Println("image bounds:", image.Bounds())
			(*chart).ChartUpdate(RaType, chartData)

			// update summary
			(*Summary).UpdateUI()

			// update slider card
			slider.rangeSlider.UpdateValues(0, float64(len(*chartData)-1), 0, float64(len(*chartData)-1))
			slider.update()
			slider.sliderCard.Hidden = false

			// slider update chart image btn
			slider.chartUpdateBtn.OnTapped = func() {
				slider.BuildSliderChartData()
				slider.UpdateChartImage(RaType, chart)

			}

			// slider reset chart image
			slider.chartResetBtn.OnTapped = func() {
				slider.ResetChartImage(RaType, chart)
			}

		}, w)

		// resize the dialog size
		RA_Dialog.Resize(fyne.Size{Width: 800, Height: 600})

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
