package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
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

	// input Vars
	iv := ntPinger.InputVars{}
	iv.Type = "dns"
	iv.Count = 0
	iv.Dns_Protocol = "udp"
	iv.Timeout = 2
	iv.Interval = 1
	iv.DestHost = "8.8.8.8"
	iv.Dns_query = "www.packetstreams.net"

	// dnsPingAddBtn action
	dnsPingAddBtn.OnTapped = func() {
		go DnsAddPingRow(&indexPing, &iv, dnsTableBody)
	}

	// Return your DNS ping interface components here
	return DNSMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: Add Ping Row
func DnsAddPingRow(indexPing *int, inputVars *ntPinger.InputVars, dnsTableBody *fyne.Container) {

	// ResultGenerateDNS()
	myDnsPing := dnsObject{}
	myDnsPing.Initial()

	// update index
	myDnsPing.DnsGUI.Index.Object.(*widget.Label).Text = strconv.Itoa(*indexPing)
	*indexPing++

	// update table body
	dnsTableBody.Add(myDnsPing.DnsGUI.DnsTableRow)
	dnsTableBody.Refresh()

	// update the close btn
	myDnsPing.DnsGUI.CloseBtn.OnTapped = func() {
		dnsTableBody.Remove(myDnsPing.DnsGUI.DnsTableRow)
		dnsTableBody.Refresh()
	}

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
		case pkt, ok := <-p.ProbeChan:

			if !ok {
				loopClose = true
				break // break select, bypass "outputChan <- pkt"
			}
			myDnsPing.DnsGUI.UpdateRow(&pkt)
			myDnsPing.UpdateChartData(&pkt)

		case err := <-errChan:
			logger.Println(err)
			return
		}
	}
}
