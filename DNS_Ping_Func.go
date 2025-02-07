package main

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt/pkg/ntTEST"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

// ******* struct dnsGUIRow ********

type dnsGUIRow struct {
	verticalSeparator *canvas.Rectangle
	Index             pingCell
	Seq               pingCell
	Status            pingCell
	Resolver          pingCell
	Query             pingCell
	Response          pingCell
	RTT               pingCell
	SendTime          pingCell
	Fail              pingCell
	MinRTT            pingCell
	MaxRTT            pingCell
	AvgRTT            pingCell
	ChartBtn          *widget.Button
	CloseBtn          *widget.Button
	Action            pingCell
	DnsTableRow       *fyne.Container
}

func (d *dnsGUIRow) Initial() {

	chartIcon := theme.NewThemedResource(resourceChartSvg)
	d.ChartBtn = widget.NewButtonWithIcon("", chartIcon, func() {})
	d.CloseBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.CloseBtn.Importance = widget.WarningImportance

	d.Action.Label = "Action"
	d.Action.Length = 80
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(2), d.ChartBtn, d.CloseBtn)

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

	d.SendTime.Label = "SendTime"
	d.SendTime.Length = 160
	d.SendTime.Object = widget.NewLabel("--")

	d.Fail.Label = "Fail"
	d.Fail.Length = 80
	d.Fail.Object = widget.NewLabel("--")

	d.MinRTT.Label = "MinRTT"
	d.MinRTT.Length = 90
	d.MinRTT.Object = widget.NewLabel("--")

	d.MaxRTT.Label = "MaxRTT"
	d.MaxRTT.Length = 90
	d.MaxRTT.Object = widget.NewLabel("--")

	d.AvgRTT.Label = "AvgRTT"
	d.AvgRTT.Length = 90
	d.AvgRTT.Object = widget.NewLabel("--")

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
		container.NewGridWrap(fyne.NewSize(float32(d.Resolver.Length), 30), d.Resolver.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Query.Length), 30), d.Query.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response.Length), 30), d.Response.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), container.NewCenter(d.RTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.SendTime.Length), 30), container.NewCenter(d.SendTime.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), container.NewCenter(d.Fail.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MinRTT.Length), 30), container.NewCenter(d.MinRTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MaxRTT.Length), 30), container.NewCenter(d.MaxRTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), container.NewCenter(d.AvgRTT.Object)),
	)
	d.DnsTableRow = container.New(layout.NewVBoxLayout(),
		row,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
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
		container.NewGridWrap(fyne.NewSize(float32(d.SendTime.Length), 30), widget.NewLabelWithStyle(d.SendTime.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), widget.NewLabelWithStyle(d.Fail.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MinRTT.Length), 30), widget.NewLabelWithStyle(d.MinRTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MaxRTT.Length), 30), widget.NewLabelWithStyle(d.MaxRTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), widget.NewLabelWithStyle(d.AvgRTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)
	headerRow := container.New(layout.NewVBoxLayout(),
		header,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
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

	// Resolver
	d.Resolver.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).DestHost
	d.Resolver.Object.(*widget.Label).Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	d.Resolver.Object.(*widget.Label).Refresh()

	// Query
	d.Query.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).Dns_query
	d.Query.Object.(*widget.Label).Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	d.Query.Object.(*widget.Label).Refresh()

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

	// SendTime
	d.SendTime.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).SendTime.Format("15:04:05")
	d.SendTime.Object.(*widget.Label).Refresh()

	// Fail Rate
	d.Fail.Object.(*widget.Label).Text = fmt.Sprintf("%.2f%%", (*p).(*ntPinger.PacketDNS).PacketLoss*100)
	d.Fail.Object.(*widget.Label).Refresh()

	// MinRTT
	d.MinRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).MinRtt.String()
	d.MinRTT.Object.(*widget.Label).Refresh()

	// MaxRTT
	d.MaxRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketDNS).MaxRtt.String()
	d.MaxRTT.Object.(*widget.Label).Refresh()

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

func ResultGenerateDNS() {

	count := 0
	Type := "dns"
	dnsSlide := []*ntPinger.PacketDNS{}

	// channel - probeChan: receiving results from probing
	// probeChan will be closed by the ResultGenerate()
	probeChan := make(chan ntPinger.Packet, 1)

	go ntTEST.ResultGenerate(count, Type, &probeChan)

	//

	// start Generating Test result
	for pkt := range probeChan {
		dnsSlide = append(dnsSlide, pkt.(*ntPinger.PacketDNS))
		fmt.Println(pkt.(*ntPinger.PacketDNS).RTT)
	}

	fmt.Println("\n--- ntTEST Testing Completed ---")
}
