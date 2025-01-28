package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

			// Get Result Analysis File Type
			RaType := records[1][0]

			appenPacket(inputResultPackets, RaType, &records)

			for _, pk := range *inputResultPackets {
				fmt.Println(*(pk.(*ntPinger.PacketDNS)))
			}

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

// func: Appen Packet Slide
func appenPacket(inputResultPackets *[]ntPinger.Packet, RaType string, records *[][]string) {
	switch RaType {
	case "dns":
		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketDNS
			p.Type = packet[0]
			p.Seq, _ = strconv.Atoi(packet[1])
			p.Status, _ = strconv.ParseBool(packet[2])
			p.DestAddr = packet[3]
			p.DestHost = packet[3]
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05", packet[9]+" "+packet[10])
			p.RTT, _ = parseCustomDuration(packet[8] + "ms")
			p.Dns_query = packet[4]
			p.Dns_queryType = packet[6]
			p.Dns_protocol = packet[7]
			p.Dns_response = packet[5]
			p.PacketsSent, _ = strconv.Atoi(packet[11])
			p.PacketsRecv, _ = strconv.Atoi(packet[12])
			p.MinRtt, _ = parseCustomDuration(packet[14])
			p.MaxRtt, _ = parseCustomDuration(packet[16])
			p.AvgRtt, _ = parseCustomDuration(packet[15])
			index_AdditionalInfo := 17
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[17]
			} else {
				p.AdditionalInfo = ""
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)

		}
	}
}

// Parse stirng (ms) to time.duration
func parseCustomDuration(input string) (time.Duration, error) {
	// Remove the "ms" suffix
	if !strings.HasSuffix(input, "ms") {
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}
	valueStr := strings.TrimSuffix(input, "ms")

	// Parse the numeric part
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", valueStr)
	}

	// Convert milliseconds to nanoseconds (time.Duration's base unit)
	duration := time.Duration(value * float64(time.Millisecond))
	return duration, nil
}
