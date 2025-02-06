package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func DNSPingContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// index
	indexPing := 1

	// ** Add-Button Card **
	dnsPingAddBtn := widget.NewButtonWithIcon("Add DNS Ping", theme.ContentAddIcon(), func() {})
	dnsPingAddBtn.Importance = widget.HighImportance
	dnsPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, dnsPingAddBtn, nil), dnsPingAddBtn)
	dnsPingAddBtncard := widget.NewCard("", "", dnsPingAddBtnContainer)

	// ** Table Container **
	dnsHeader := dnsGUIRow{}
	dnsHeader.Initial()
	dnsHeaderRow := dnsHeader.GenerateHeaderRow()

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
		myDnsPing := dnsObject{}
		myDnsPing.Initial()

		// update index
		myDnsPing.DnsGUI.Index.Object.(*widget.Label).Text = strconv.Itoa(indexPing)
		indexPing++

		// update table body
		dnsTableBody.Add(myDnsPing.DnsGUI.DnsTableRow)
		dnsTableBody.Refresh()

		// update the close btn
		myDnsPing.DnsGUI.CloseBtn.OnTapped = func() {
			dnsTableBody.Remove(myDnsPing.DnsGUI.DnsTableRow)
			dnsTableBody.Refresh()
		}
	}

	// Return your DNS ping interface components here
	return DNSMainContainerOuter // Temporary empty container, replace with your actual UI
}
