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
)

func DNSPingContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// index
	ntGlobal.dnsIndex = 1

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
	ntGlobal.dnsTable = dnsTableBody

	dnsTableScroll := container.NewScroll(ntGlobal.dnsTable)
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
	iv.Dns_query = "netflix.com"

	// dnsPingAddBtn action
	dnsPingAddBtn.OnTapped = func() {
		NewTest(a, "dns", ntGlobal.dnsTable)
		//go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable)
	}

	// Return your DNS ping interface components here
	return DNSMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: Add Ping Row
func DnsAddPingRow(a fyne.App, indexPing *int, inputVars *ntPinger.InputVars, dnsTableBody *fyne.Container) {

	// ResultGenerateDNS()
	myDnsPing := dnsObject{}
	myDnsPing.Initial()

	// update index
	myPingIndex := strconv.Itoa(*indexPing)

	myDnsPing.DnsGUI.Index.Object.(*widget.Label).Text = myPingIndex
	*indexPing++

	// Update Resolver
	myDnsPing.DnsGUI.Resolver.Object.(*widget.Label).Text = TruncateString(inputVars.DestHost, 22)

	// Update DNS Query
	myDnsPing.DnsGUI.Query.Object.(*widget.Label).Text = TruncateString(inputVars.Dns_query, 25)

	// Update StartTime
	myDnsPing.DnsGUI.StartTime.Object.(*widget.Label).Text = time.Now().Format("2006-01-02 15:04:05 MST")

	// update table body
	dnsTableBody.Add(myDnsPing.DnsGUI.DnsTableRow)
	dnsTableBody.Refresh()

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
		go DnsAddPingRow(a, indexPing, inputVars, dnsTableBody)
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

		// harvest the errChan input
		case err := <-errChan:
			logger.Println(err)
			return
		}
	}

	// update test table when test is closed

	// deal with the recordingChan when test is closed

}
