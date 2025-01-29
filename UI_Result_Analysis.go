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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	ntHttp "github.com/djian01/nt/pkg/cmd/http"
	"github.com/djian01/nt/pkg/ntPinger"

	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

func ResultAnalysisContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// Initial slides
	inputResultPackets := []ntPinger.Packet{}
	chartData := []ntchart.ChartPoint{}

	// Input Result CSV File card
	inputResultCSVFilePath := widget.NewEntry()
	inputResultCSVFilePath.SetPlaceHolder("please input the file path or press the right button to select the Result CSV file")
	inputResultCSVFileButton := widget.NewButton("Select the Result CSV File", func() {})
	inputResultCSVFileButton.Importance = widget.WarningImportance
	inputNSXConfigContainer := container.New(layout.NewBorderLayout(nil, nil, nil, inputResultCSVFileButton), inputResultCSVFilePath, inputResultCSVFileButton)
	inputResultCSVFileCard := widget.NewCard("", "Input the existing Result CSV File", inputNSXConfigContainer)

	// Chart Card (Place Holder)
	chartImage := canvas.NewImageFromResource(nil)
	chartImage.FillMode = canvas.ImageFillContain
	chartImage.SetMinSize(fyne.NewSize(400, 400))

	// Create a grid with the placeholder
	chartContainer := container.New(layout.NewBorderLayout(nil, nil, nil, nil), chartImage)
	chartCard := widget.NewCard("", "", chartContainer)

	//// Main Container
	spaceHolder := widget.NewLabel("                     ")
	RaMainContainerInner := container.New(layout.NewVBoxLayout(), inputResultCSVFileCard, chartCard)
	RaMainContainerOuter := container.New(layout.NewBorderLayout(spaceHolder, spaceHolder, spaceHolder, spaceHolder), spaceHolder, RaMainContainerInner)

	// Input NSX Config File BTN
	inputResultCSVFileButton.OnTapped = OpenResultCSVFile(w, &inputResultPackets, &chartData, chartImage, chartContainer)

	// Return your result analysis interface components here
	return RaMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: OpenResultCSVFile
func OpenResultCSVFile(w fyne.Window, inputResultPackets *[]ntPinger.Packet, chartData *[]ntchart.ChartPoint, chartImage *canvas.Image, chartContainer *fyne.Container) func() {
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

			appenPacket(inputResultPackets, RaType, &records, chartData)

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
			image := ntchart.CreateChart("My Chart", chartData, 0)

			// verify the image.Bounds(), e.g. image bounds: (0,0)-(1024,512) is good
			// fmt.Println("image bounds:", image.Bounds())

			chartImage.Image = image
			chartImage.FillMode = canvas.ImageFillStretch
			chartImage.Refresh()

			chartContainer.Refresh()

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
func appenPacket(inputResultPackets *[]ntPinger.Packet, RaType string, records *[][]string, chartData *[]ntchart.ChartPoint) {

	var chartPoint ntchart.ChartPoint

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
			chartPoint.Status = p.Status
			p.DestAddr = packet[3]
			p.DestHost = packet[3]
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05", packet[9]+" "+packet[10])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[8] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
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
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)

		}
	case "http":

		RAHttpVar, _ := ntHttp.ParseURL((*records)[1][4])

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketHTTP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestHost = RAHttpVar.Hostname
			p.DestPort = RAHttpVar.Port
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05", packet[8]+" "+packet[9])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[7] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.Http_path = RAHttpVar.Path
			p.Http_scheme = RAHttpVar.Scheme
			p.Http_response_code, _ = strconv.Atoi(packet[5])
			p.Http_response = packet[6]
			p.Http_method = packet[3]
			p.PacketsSent, _ = strconv.Atoi(packet[10])
			p.PacketsRecv, _ = strconv.Atoi(packet[11])
			p.MinRtt, _ = parseCustomDuration(packet[13])
			p.MaxRtt, _ = parseCustomDuration(packet[15])
			p.AvgRtt, _ = parseCustomDuration(packet[14])
			index_AdditionalInfo := 16
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}

	case "tcp":

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketTCP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestAddr = packet[4]
			p.DestHost = packet[3]
			p.DestPort, _ = strconv.Atoi(packet[5])
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05", packet[8]+" "+packet[9])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[7] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.PayLoadSize, _ = strconv.Atoi(packet[6])
			p.PacketsSent, _ = strconv.Atoi(packet[10])
			p.PacketsRecv, _ = strconv.Atoi(packet[11])
			p.MinRtt, _ = parseCustomDuration(packet[13])
			p.MaxRtt, _ = parseCustomDuration(packet[15])
			p.AvgRtt, _ = parseCustomDuration(packet[14])
			index_AdditionalInfo := 16
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}

	case "icmp":

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketICMP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestAddr = packet[4]
			p.DestHost = packet[3]
			p.PayLoadSize, _ = strconv.Atoi(packet[5])
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05", packet[7]+" "+packet[8])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[6] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.PacketsSent, _ = strconv.Atoi(packet[9])
			p.PacketsRecv, _ = strconv.Atoi(packet[10])
			p.MinRtt, _ = parseCustomDuration(packet[12])
			p.MaxRtt, _ = parseCustomDuration(packet[14])
			p.AvgRtt, _ = parseCustomDuration(packet[13])
			index_AdditionalInfo := 15
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}
	}
}

// Parse a duration string with "ms" (milliseconds) or "s" (seconds) to time.Duration
func parseCustomDuration(input string) (time.Duration, error) {

	var multiplier float64

	// Check the suffix and set the multiplier accordingly
	switch {
	case strings.HasSuffix(input, "ms"):
		multiplier = float64(time.Millisecond)
		input = strings.TrimSuffix(input, "ms") // Remove the "ms" suffix
	case strings.HasSuffix(input, "s"):
		multiplier = float64(time.Second)
		input = strings.TrimSuffix(input, "s") // Remove the "s" suffix
	default:
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", input)
	}

	// Use the multiplier to compute the duration
	duration := time.Duration(value * multiplier)
	return duration, nil
}
