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
)

type dnsPingRow struct {
	FailCount   int
	results     []ntPinger.PacketDNS
	Seq         *widget.Label
	Status      *canvas.Image
	Resolver    *widget.Label
	Query       *widget.Label
	Response    *widget.Label
	RTT         *widget.Label
	SendTime    *widget.Label
	AddInfo     *widget.Label
	Fail        *widget.Label
	MinRTT      *widget.Label
	MaxRTT      *widget.Label
	AvgRTT      *widget.Label
	ChartBtn    *widget.Button
	CloseBtn    *widget.Button
	Action      *fyne.Container
	DnsTableRow *fyne.Container
}

func (d *dnsPingRow) Initial() {

	// initial fail count
	d.FailCount = 0

	// initial widgets
	d.Seq = widget.NewLabel("")
	d.Status = canvas.NewImageFromResource(theme.MediaRecordIcon())
	d.Status.FillMode = canvas.ImageFillContain
	d.Resolver = widget.NewLabel("")
	d.Query = widget.NewLabel("")
	d.Response = widget.NewLabel("")
	d.RTT = widget.NewLabel("")
	d.SendTime = widget.NewLabel("")
	d.AddInfo = widget.NewLabel("")
	d.Fail = widget.NewLabel("")
	d.MinRTT = widget.NewLabel("")
	d.MaxRTT = widget.NewLabel("")
	d.AvgRTT = widget.NewLabel("")
	d.ChartBtn = widget.NewButton("Chart", func() {})
	d.CloseBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.Action = container.New(layout.NewGridLayoutWithColumns(2), d.ChartBtn, d.CloseBtn)

	// table row
	row := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(50, 30), d.Seq),
		container.NewGridWrap(fyne.NewSize(65, 30), d.Status),
		container.NewGridWrap(fyne.NewSize(145, 30), d.Resolver),
		container.NewGridWrap(fyne.NewSize(160, 30), d.Query),
		container.NewGridWrap(fyne.NewSize(160, 30), d.Response),
		container.NewGridWrap(fyne.NewSize(75, 30), d.RTT),
		container.NewGridWrap(fyne.NewSize(100, 30), d.SendTime),
		container.NewGridWrap(fyne.NewSize(100, 30), d.AddInfo),
		container.NewGridWrap(fyne.NewSize(60, 30), d.Fail),
		container.NewGridWrap(fyne.NewSize(75, 30), d.MinRTT),
		container.NewGridWrap(fyne.NewSize(80, 30), d.MaxRTT),
		container.NewGridWrap(fyne.NewSize(75, 30), d.AvgRTT),
		container.NewGridWrap(fyne.NewSize(100, 30), d.Action),
	)
	d.DnsTableRow = container.New(layout.NewVBoxLayout(),
		row,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
	)
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
