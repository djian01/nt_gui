package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func DNSPingContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// ** Add-Button Card **
	dnsPingAddBtn := widget.NewButtonWithIcon("Add DNS Ping", theme.ContentAddIcon(), func() {})
	dnsPingAddBtn.Importance = widget.HighImportance
	dnsPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, dnsPingAddBtn, nil), dnsPingAddBtn)
	dnsPingAddBtncard := widget.NewCard("", "", dnsPingAddBtnContainer)

	// ** Table Container **
	dnsHeader := []pingCell{
		{Label: "Seq", Length: 50},
		{Label: "Status", Length: 65},
		{Label: "Resolver", Length: 145},
		{Label: "Query", Length: 160},
		{Label: "Response", Length: 160},
		{Label: "RTT", Length: 75},
		{Label: "Send_Time", Length: 100},
		{Label: "Add_Info", Length: 100},
		{Label: "Fail", Length: 60},
		{Label: "Min_RTT", Length: 75},
		{Label: "Max_RTT", Length: 80},
		{Label: "Avg_RTT", Length: 75},
		{Label: "Action", Length: 100},
	}

	dnsHeaderTable := widget.NewTable(
		// callback - length
		func() (int, int) {
			return 1, len(dnsHeader)
		},
		// callback - CreateCell
		func() fyne.CanvasObject {
			// the basic Cell Width is defined by the string length below
			item := widget.NewLabel("")
			item.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			return item
		},
		// callback - UpdateCell
		//// cell: contains Row, Col int for a cell, item: the CanvasObject from CreateCell Callback
		func(id widget.TableCellID, item fyne.CanvasObject) {
			// Header row
			item.(*widget.Label).TextStyle.Bold = true
			item.(*widget.Label).SetText(dnsHeader[id.Col].Label)
		},
	)

	// set width for all colunms width for loop
	for i := 0; i < len(dnsHeader); i++ {
		dnsHeaderTable.SetColumnWidth(i, float32(dnsHeader[i].Length))
	}

	dnsHeaderRow := container.New(layout.NewVBoxLayout(),
		dnsHeaderTable,
		canvas.NewLine(color.RGBA{200, 200, 200, 255}),
	)

	dnsTableBody := container.New(layout.NewVBoxLayout())
	dnsTableScroll := container.NewScroll(dnsTableBody)
	dnsTableContainer := container.New(layout.NewBorderLayout(dnsHeaderRow, nil, nil, nil), dnsHeaderRow, dnsTableScroll)

	// ** Table Card **
	dnsTableCard := widget.NewCard("", "", dnsTableContainer)

	// ** Main Container **
	DNSSpaceHolder := widget.NewLabel("    ")
	DNSMainContainerInner := container.New(layout.NewBorderLayout(dnsPingAddBtncard, nil, nil, nil), dnsPingAddBtncard, dnsTableCard)
	DNSMainContainerOuter := container.New(layout.NewBorderLayout(DNSSpaceHolder, DNSSpaceHolder, DNSSpaceHolder, DNSSpaceHolder), DNSSpaceHolder, DNSMainContainerInner)

	// dnsPingAddBtn action
	dnsPingAddBtn.OnTapped = func() {
		// ResultGenerateDNS()
		myD1 := dnsPingRow{}
		myD1.Initial()
		dnsTableBody.Add(myD1.DnsTableRow)
		dnsTableBody.Refresh()
	}

	// Return your DNS ping interface components here
	return DNSMainContainerOuter // Temporary empty container, replace with your actual UI
}
