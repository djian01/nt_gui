package main

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

// ******* struct dnsGUIRow ********

type dnsGUIRow struct {
	Index       pingCell
	Seq         pingCell
	Status      pingCell
	Resolver    pingCell // fixed
	Query       pingCell // fixed
	Response    pingCell
	RTT         pingCell
	StartTime   pingCell // fixed
	Fail        pingCell
	AvgRTT      pingCell
	Recording   pingCell
	ChartBtn    *widget.Button
	StopBtn     *widget.Button
	ReplayBtn   *widget.Button
	CloseBtn    *widget.Button
	Action      pingCell
	DnsTableRow *fyne.Container
}

func (d *dnsGUIRow) Initial() {

	chartIcon := theme.NewThemedResource(resourceChartSvg)
	d.ChartBtn = widget.NewButtonWithIcon("", chartIcon, func() {})

	d.StopBtn = widget.NewButtonWithIcon("", theme.MediaStopIcon(), func() {})
	d.StopBtn.Importance = widget.DangerImportance

	d.ReplayBtn = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {})
	d.ReplayBtn.Importance = widget.HighImportance
	d.ReplayBtn.Disable()

	d.CloseBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.CloseBtn.Importance = widget.WarningImportance
	d.CloseBtn.Disable()

	d.Action.Label = "Action"
	d.Action.Length = 110
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(4), d.ChartBtn, d.StopBtn, d.ReplayBtn, d.CloseBtn)

	d.Index.Label = "Index"
	d.Index.Length = 50
	d.Index.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	d.Seq.Label = "Seq"
	d.Seq.Length = 50
	d.Seq.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	d.Status.Label = "Status"
	d.Status.Length = 65
	d.Status.Object = canvas.NewText("--", color.Black)

	d.Resolver.Label = "Resolver"
	d.Resolver.Length = 160
	d.Resolver.Object = widget.NewLabel("--")

	d.Query.Label = "Query"
	d.Query.Length = 180
	d.Query.Object = widget.NewLabel("--")

	d.Response.Label = "Response"
	d.Response.Length = 180
	d.Response.Object = widget.NewLabel("--")

	d.RTT.Label = "RTT"
	d.RTT.Length = 90
	d.RTT.Object = widget.NewLabel("--")

	d.StartTime.Label = "StartTime"
	d.StartTime.Length = 190
	d.StartTime.Object = widget.NewLabel("--")

	d.Fail.Label = "Fail"
	d.Fail.Length = 80
	d.Fail.Object = widget.NewLabel("--")

	d.AvgRTT.Label = "AvgRTT"
	d.AvgRTT.Length = 90
	d.AvgRTT.Object = widget.NewLabel("--")

	d.Recording.Label = "Recording"
	d.Recording.Length = 80
	d.Recording.Object = widget.NewLabel("OFF")

	// table row
	row := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), d.Action.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), container.NewCenter(d.Index.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), container.NewCenter(d.Seq.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), container.NewCenter(d.Status.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Resolver.Length), 30), container.NewCenter(d.Resolver.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Query.Length), 30), container.NewCenter(d.Query.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response.Length), 30), d.Response.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), container.NewCenter(d.RTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), container.NewCenter(d.StartTime.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), container.NewCenter(d.Fail.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), container.NewCenter(d.AvgRTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Recording.Length), 30), container.NewCenter(d.Recording.Object)),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 2)) // Adjust width & thickness

	d.DnsTableRow = container.New(layout.NewVBoxLayout(),
		row,
		thickLine,
	)
}

func (d *dnsGUIRow) GenerateHeaderRow() *fyne.Container {

	// table row
	header := container.New(layout.NewHBoxLayout(),

		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), widget.NewLabelWithStyle(d.Action.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), widget.NewLabelWithStyle(d.Index.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), widget.NewLabelWithStyle(d.Seq.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), widget.NewLabelWithStyle(d.Status.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Resolver.Length), 30), widget.NewLabelWithStyle(d.Resolver.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Query.Length), 30), widget.NewLabelWithStyle(d.Query.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response.Length), 30), widget.NewLabelWithStyle(d.Response.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), widget.NewLabelWithStyle(d.RTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), widget.NewLabelWithStyle(d.StartTime.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), widget.NewLabelWithStyle(d.Fail.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), widget.NewLabelWithStyle(d.AvgRTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Recording.Length), 30), widget.NewLabelWithStyle(d.Recording.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 3)) // Adjust width & thickness

	headerRow := container.New(layout.NewVBoxLayout(),
		header,
		thickLine,
	)

	return headerRow
}

func (d *dnsGUIRow) UpdateRow(p *ntPinger.Packet) {

	// seq
	d.Seq.Object.(*widget.Label).Text = strconv.Itoa((*p).(*ntPinger.PacketDNS).Seq)
	d.Seq.Object.(*widget.Label).Refresh()

	// status
	d.Status.Object.(*canvas.Text).TextStyle.Bold = true
	d.Status.Object.(*canvas.Text).TextSize = 15
	if (*p).(*ntPinger.PacketDNS).Status {
		d.Status.Object.(*canvas.Text).Text = "Success"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{0, 128, 0, 255}
	} else {
		d.Status.Object.(*canvas.Text).Text = "Fail"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{255, 0, 0, 255}
	}
	d.Status.Object.(*canvas.Text).Refresh()

	// Response
	d.Response.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).Dns_response
	d.Response.Object.(*widget.Label).Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	d.Response.Object.(*widget.Label).Refresh()

	// RTT
	if (*p).(*ntPinger.PacketDNS).Status {
		d.RTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).RTT.String()
	} else {
		d.RTT.Object.(*widget.Label).Text = "--"
	}
	d.RTT.Object.(*widget.Label).Refresh()

	// Fail Rate
	d.Fail.Object.(*widget.Label).Text = fmt.Sprintf("%.2f%%", (*p).(*ntPinger.PacketDNS).PacketLoss*100)
	d.Fail.Object.(*widget.Label).Refresh()

	// AvgRTT
	d.AvgRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).AvgRtt.String()
	d.AvgRTT.Object.(*widget.Label).Refresh()
}

// ******* struct dnsObject ********

type dnsObject struct {
	FailCount int
	ChartData []ntchart.ChartPoint
	DnsGUI    dnsGUIRow
}

func (d *dnsObject) Initial() {
	// initial fail count
	d.FailCount = 0

	// ChartData
	d.ChartData = []ntchart.ChartPoint{}

	// Dns GUI
	d.DnsGUI = dnsGUIRow{}
	d.DnsGUI.Initial()
}

func (d *dnsObject) UpdateChartData(pkt *ntPinger.Packet) {
	d.ChartData = append(d.ChartData, ntchart.ConvertFromPacketToChartPoint(*pkt))
}

func (d *dnsObject) DisplayChartDataTerminal() {
	for _, d := range d.ChartData {
		fmt.Println(d)
	}
}

// func: Add Ping Row
func DnsAddPingRow(a fyne.App, indexPing *int, inputVars *ntPinger.InputVars, dnsTableBody *fyne.Container, recording bool) {

	// ResultGenerateDNS()
	myDnsPing := dnsObject{}
	myDnsPing.Initial()

	// update index
	myPingIndex := strconv.Itoa(*indexPing)

	myDnsPing.DnsGUI.Index.Object.(*widget.Label).Text = myPingIndex
	myDnsPing.DnsGUI.Index.Object.(*widget.Label).Refresh()
	*indexPing++

	// Update Resolver
	myDnsPing.DnsGUI.Resolver.Object.(*widget.Label).Text = TruncateString(inputVars.DestHost, 22)
	myDnsPing.DnsGUI.Resolver.Object.(*widget.Label).Refresh()

	// Update DNS Query
	myDnsPing.DnsGUI.Query.Object.(*widget.Label).Text = TruncateString(inputVars.Dns_query, 25)
	myDnsPing.DnsGUI.Query.Object.(*widget.Label).Refresh()

	// Update StartTime
	myDnsPing.DnsGUI.StartTime.Object.(*widget.Label).Text = time.Now().Format("2006-01-02 15:04:05 MST")
	myDnsPing.DnsGUI.StartTime.Object.(*widget.Label).Refresh()

	// update table body
	dnsTableBody.Add(myDnsPing.DnsGUI.DnsTableRow)
	dnsTableBody.Refresh()

	// update recording
	if recording {
		myDnsPing.DnsGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		myDnsPing.DnsGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}
	myDnsPing.DnsGUI.Recording.Object.(*widget.Label).Refresh()

	// Add New Entry to DB History

	// ** start ntPinger Probe **

	// Channel - error (for Go Routines)
	errChan := make(chan error, 1)
	defer close(errChan)

	// Start Ping Main Command, manually input display Len
	p, err := ntPinger.NewPinger(*inputVars)

	if err != nil {
		fmt.Println(err)
		logger.Println(err)
		return
	}

	// OnTapped Func - Chart btn
	myDnsPing.DnsGUI.ChartBtn.OnTapped = func() {
		myCmd := NtCmdGenerator(true, *inputVars)
		fmt.Println(myCmd)
	}

	// OnTapped Func - Stop btn
	myDnsPing.DnsGUI.StopBtn.OnTapped = func() {
		p.PingerEnd = true
		time.Sleep(200 * time.Millisecond) // wait for the test to stop

		myDnsPing.DnsGUI.StopBtn.Disable()
		myDnsPing.DnsGUI.CloseBtn.Enable()
		myDnsPing.DnsGUI.ReplayBtn.Enable()

		myDnsPing.DnsGUI.Status.Object.(*canvas.Text).Text = "Stop"
		myDnsPing.DnsGUI.Status.Object.(*canvas.Text).Color = color.RGBA{165, 42, 42, 255}
		myDnsPing.DnsGUI.Status.Object.(*canvas.Text).Refresh()

	}

	// OnTapped Func - Replay btn
	myDnsPing.DnsGUI.ReplayBtn.OnTapped = func() {
		// re-launch a new go routine for DnsAddPingRow with the same InputVar
		go DnsAddPingRow(a, indexPing, inputVars, dnsTableBody, recording)
	}

	// OnTapped Func - close btn
	myDnsPing.DnsGUI.CloseBtn.OnTapped = func() {
		dnsTableBody.Remove(myDnsPing.DnsGUI.DnsTableRow)
		dnsTableBody.Refresh()
	}

	// start ping go routing
	go p.Run(errChan)

	// harvest the result
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
			p.PingerEnd = true
			loopClose = true
			//fmt.Printf("Closing Testing: %s\n", myPingIndex)

		// harvest the Probe results
		case pkt, ok := <-p.ProbeChan:

			// if p.ProbeChan is closed, exit
			if !ok {
				loopClose = true
				break // break select, bypass following code in the same case
			}
			myDnsPing.DnsGUI.UpdateRow(&pkt)
			myDnsPing.UpdateChartData(&pkt)

			// Add test result entry to DB is "recording" is "ON"

		// harvest the errChan input
		case err := <-errChan:
			logger.Println(err)
			return
		}
	}

	// update test table when test is closed

	// deal with the recordingChan when test is closed

}
