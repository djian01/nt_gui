package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	Summary := Summary{}

	// Input Result CSV File card
	inputResultCSVFilePath := widget.NewEntry()
	inputResultCSVFilePath.SetPlaceHolder("please press the right button to select the Result CSV file")
	inputResultCSVFileButton := widget.NewButton("Select the Result CSV File", func() {})
	inputResultCSVFileButton.Importance = widget.WarningImportance
	inputNSXConfigContainer := container.New(layout.NewBorderLayout(nil, nil, nil, inputResultCSVFileButton), inputResultCSVFilePath, inputResultCSVFileButton)
	inputResultCSVFileCard := widget.NewCard("", "Input the existing Result CSV File", inputNSXConfigContainer)

	// Summary Card
	summaryUI := SummaryUI{}
	summaryUI.Initial()
	summaryCard := summaryUI.CreateCard()

	// Chart Card (Place Holder)
	chartImage := canvas.NewImageFromResource(nil)
	chartImage.FillMode = canvas.ImageFillContain
	chartImage.SetMinSize(fyne.NewSize(400, 400))

	// Create a grid with the placeholder
	chartContainer := container.New(layout.NewBorderLayout(nil, nil, nil, nil), chartImage)
	chartCard := widget.NewCard("", "", chartContainer)

	//// Main Container
	spaceHolder := widget.NewLabel("                     ")
	RaMainContainerInner := container.New(layout.NewVBoxLayout(), inputResultCSVFileCard, summaryCard, chartCard)
	RaMainContainerOuter := container.New(layout.NewBorderLayout(spaceHolder, spaceHolder, spaceHolder, spaceHolder), spaceHolder, RaMainContainerInner)

	// Input NSX Config File BTN
	inputResultCSVFileButton.OnTapped = OpenResultCSVFile(w, &inputResultPackets, &chartData, chartImage, chartContainer, &Summary, &summaryUI, inputResultCSVFilePath)

	// Return your result analysis interface components here
	return RaMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: OpenResultCSVFile
func OpenResultCSVFile(w fyne.Window, inputResultPackets *[]ntPinger.Packet, chartData *[]ntchart.ChartPoint, chartImage *canvas.Image, chartContainer *fyne.Container, Summary *Summary, summaryUI *SummaryUI, inputResultCSVFilePath *widget.Entry) func() {
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

			// for _, pk := range *inputResultPackets {

			// 	switch RaType {
			// 	case "dns":
			// 		fmt.Println(*(pk.(*ntPinger.PacketDNS)))
			// 	case "http":
			// 		fmt.Println(*(pk.(*ntPinger.PacketHTTP)))
			// 	case "tcp":
			// 		fmt.Println(*(pk.(*ntPinger.PacketTCP)))
			// 	case "icmp":
			// 		fmt.Println(*(pk.(*ntPinger.PacketICMP)))
			// 	}
			// }

			// for _, chartPoint := range *chartData {
			// 	fmt.Printf("%v, %v, %v\n", chartPoint.Status, chartPoint.XValues, chartPoint.YValues)
			// }

			// Create an image Chart
			// verify the image.Bounds(), e.g. image bounds: (0,0)-(1024,512) is good. code -> fmt.Println("image bounds:", image.Bounds())
			image := ntchart.CreateChart(RaType, chartData, 0)
			chartImage.Image = image
			chartImage.FillMode = canvas.ImageFillStretch
			chartImage.Refresh()

			chartContainer.Refresh()

			// update summary
			summaryUI.typeEntry.SetText((*Summary).Type)
			summaryUI.destHostEntry.SetText((*Summary).DestHost)

			summaryUI.startTimeEntry.SetText((*Summary).StartTime.Format(("2006-01-02 15:04:05 MST")))
			summaryUI.endTimeEntry.SetText((*Summary).EndTime.Format(("2006-01-02 15:04:05 MST")))
			summaryUI.packetSentEntry.SetText(strconv.Itoa((*Summary).PacketSent))
			summaryUI.successResponseEntry.SetText(strconv.Itoa((*Summary).SuccessResponse))
			summaryUI.failRateEntry.SetText((*Summary).FailRate)
			summaryUI.minRttEntry.SetText(fmt.Sprintf("%d ms", (*Summary).MinRTT.Milliseconds()))
			summaryUI.maxRttEntry.SetText(fmt.Sprintf("%d ms", (*Summary).MaxRTT.Milliseconds()))
			summaryUI.avgRttEntry.SetText(fmt.Sprintf("%d ms", (*Summary).AvgRtt.Milliseconds()))
			summaryUI.ntCmdEntry.SetText((*Summary).ntCmd)

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
