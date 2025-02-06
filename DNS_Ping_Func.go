package main

import (
	"fmt"
	"image/color"

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
	AddInfo           pingCell
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
	//d.ChartBtn.Importance = widget.LowImportance
	d.CloseBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.CloseBtn.Importance = widget.WarningImportance

	d.Action.Label = "Action"
	d.Action.Length = 80
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(2), d.ChartBtn, d.CloseBtn)

	d.Index.Label = "Index"
	d.Index.Length = 50
	d.Index.Object = widget.NewLabelWithStyle(d.Seq.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	d.Seq.Label = "Seq"
	d.Seq.Length = 50
	d.Seq.Object = widget.NewLabel(d.Seq.Label)

	d.Status.Label = "Status"
	d.Status.Length = 65
	statusImage := canvas.NewImageFromResource(theme.MediaRecordIcon())
	statusImage.FillMode = canvas.ImageFillContain
	d.Status.Object = statusImage

	d.Resolver.Label = "Resolver"
	d.Resolver.Length = 145
	d.Resolver.Object = widget.NewLabel(d.Resolver.Label)

	d.Query.Label = "Query"
	d.Query.Length = 160
	d.Query.Object = widget.NewLabel(d.Query.Label)

	d.Response.Label = "Response"
	d.Response.Length = 160
	d.Response.Object = widget.NewLabel(d.Response.Label)

	d.RTT.Label = "RTT"
	d.RTT.Length = 75
	d.RTT.Object = widget.NewLabel(d.RTT.Label)

	d.SendTime.Label = "SendTime"
	d.SendTime.Length = 160
	d.SendTime.Object = widget.NewLabel(d.SendTime.Label)

	d.AddInfo.Label = "AddInfo"
	d.AddInfo.Length = 100
	d.AddInfo.Object = widget.NewLabel(d.AddInfo.Label)

	d.Fail.Label = "Fail"
	d.Fail.Length = 60
	d.Fail.Object = widget.NewLabel(d.Fail.Label)

	d.MinRTT.Label = "MinRTT"
	d.MinRTT.Length = 75
	d.MinRTT.Object = widget.NewLabel(d.MinRTT.Label)

	d.MaxRTT.Label = "MaxRTT"
	d.MaxRTT.Length = 80
	d.MaxRTT.Object = widget.NewLabel(d.MaxRTT.Label)

	d.AvgRTT.Label = "AvgRTT"
	d.AvgRTT.Length = 75
	d.AvgRTT.Object = widget.NewLabel(d.AvgRTT.Label)

	// table row
	row := container.New(layout.NewHBoxLayout(),

		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), d.Action.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), d.Index.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), d.Seq.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), d.Status.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Resolver.Length), 30), d.Resolver.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Query.Length), 30), d.Query.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response.Length), 30), d.Response.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), d.RTT.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.SendTime.Length), 30), d.SendTime.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AddInfo.Length), 30), d.AddInfo.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), d.Fail.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MinRTT.Length), 30), d.MinRTT.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MaxRTT.Length), 30), d.MaxRTT.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), d.AvgRTT.Object),
	)
	d.DnsTableRow = container.New(layout.NewVBoxLayout(),
		row,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
	)
}

func (d *dnsGUIRow) GenerateHeaderRow() *fyne.Container {

	// table row
	header := container.New(layout.NewHBoxLayout(),

		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), widget.NewLabelWithStyle(d.Action.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), widget.NewLabelWithStyle(d.Index.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), widget.NewLabelWithStyle(d.Seq.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), widget.NewLabelWithStyle(d.Status.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Resolver.Length), 30), widget.NewLabelWithStyle(d.Resolver.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Query.Length), 30), widget.NewLabelWithStyle(d.Query.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response.Length), 30), widget.NewLabelWithStyle(d.Response.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), widget.NewLabelWithStyle(d.RTT.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.SendTime.Length), 30), widget.NewLabelWithStyle(d.SendTime.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AddInfo.Length), 30), widget.NewLabelWithStyle(d.AddInfo.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), widget.NewLabelWithStyle(d.Fail.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MinRTT.Length), 30), widget.NewLabelWithStyle(d.MinRTT.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.MaxRTT.Length), 30), widget.NewLabelWithStyle(d.MaxRTT.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), widget.NewLabelWithStyle(d.AvgRTT.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
	)
	headerRow := container.New(layout.NewVBoxLayout(),
		header,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
	)

	return headerRow
}

type dnsObject struct {
	FailCount   int
	GuiPingChan chan ntPinger.Packet
	ChartData   []ntchart.ChartPoint
	DnsGUI      dnsGUIRow
}

func (d *dnsObject) Initial() {
	// initial fail count
	d.FailCount = 0

	// Packet Chan
	d.GuiPingChan = make(chan ntPinger.Packet, 1)

	// ChartData
	d.ChartData = []ntchart.ChartPoint{}

	// Dns GUI
	d.DnsGUI = dnsGUIRow{}
	d.DnsGUI.Initial()
}

func (d *dnsObject) UpdateChartData() {
	// range the channel
	for pkt := range d.GuiPingChan {
		d.ChartData = append(d.ChartData, ntchart.ConvertFromPacketToChartPoint(pkt))
	}

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
