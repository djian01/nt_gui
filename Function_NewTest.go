package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

// func: New Test Input
func NewTest(a fyne.App, testType string, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) {

	// Initial New Test Input Var Window
	newTestWindow := a.NewWindow(fmt.Sprintf("New %s Test", strings.ToUpper(testType)))
	newTestWindow.CenterOnScreen()

	// btns
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {})
	cancelBtn.Importance = widget.WarningImportance
	cancelBtn.OnTapped = func() {
		newTestWindow.Close()
	}
	submitBtn := widget.NewButtonWithIcon("Submit", theme.ConfirmIcon(), func() {})
	submitBtn.Importance = widget.HighImportance
	btnContainer := formCell(cancelBtn, 100, submitBtn, 100)

	// error message
	errMsg := canvas.NewText("", color.RGBA{255, 0, 0, 255})
	errMsg.TextStyle.Bold = true
	errMsg.Text = ""
	errMsgContainer := container.New(layout.NewVBoxLayout(), errMsg)

	// common container (common for all test types)
	// Interval
	intervalValue := 1
	intervalCheck := true
	intervalLabel := widget.NewLabel("Interval (s):")
	intervalEntry := widget.NewEntry()
	intervalEntry.Text = "1"
	intervalEntry.Validator = func(s string) error {
		// Convert text to an integer
		num, err := strconv.Atoi(s)
		if err != nil || num < 1 {
			intervalCheck = false
			msg := "interval should always be Int and larger than 0"
			errMsg.Text = msg
			errMsg.Refresh()
			return fmt.Errorf("validation error: %s", msg)
		} else {
			intervalCheck = true
			intervalValue = num
			errMsg.Text = ""
			errMsg.Refresh()
			return nil
		}
	}

	intervalContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), intervalEntry)
	intervalCell := formCell(intervalLabel, 100, intervalContainer, 100)

	// Timeout
	timeoutValue := 4
	timeoutCheck := true
	timeoutLabel := widget.NewLabel("Timeout (s):")
	timeoutEntry := widget.NewEntry()
	timeoutEntry.Text = "4"
	timeoutEntry.Validator = func(s string) error {
		// Convert text to an integer
		num, err := strconv.Atoi(s)
		if err != nil || num < 1 {
			timeoutCheck = false
			msg := "Timeout should always be Int and larger than 0"
			errMsg.Text = msg
			errMsg.Refresh()
			return fmt.Errorf("validation error: %s", msg)
		} else {
			timeoutCheck = true
			timeoutValue = num
			errMsg.Text = ""
			errMsg.Refresh()
			return nil
		}
	}
	timeoutCell := formCell(timeoutLabel, 100, timeoutEntry, 100)

	// recording
	recording := false
	recordingLabel := widget.NewLabel("Result Recording:")
	recordingCheck := widget.NewCheck("", func(b bool) {
		recording = b
	})
	recordingCell := formCell(recordingLabel, 150, recordingCheck, 50)

	// common Container
	commonContainerSub := formCell(intervalCell, 250, timeoutCell, 250)
	commonContainer := container.NewVBox(recordingCell, commonContainerSub)

	// Specific Vars
	specificContainer := container.NewVBox()

	switch testType {
	case "dns":
		// update the New Test Window Size
		newTestWindow.Resize(fyne.NewSize(710, 550))

		// target server
		dnsServerCheck := false
		dnsServerLabel := widget.NewLabel("DNS Server IP/Host(s):")
		dnsServerEntry := widget.NewMultiLineEntry()
		dnsServerEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 150)), dnsServerEntry)
		dnsServerContainer := container.New(layout.NewVBoxLayout(), dnsServerLabel, dnsServerEntryContainer)

		// dns query
		dnsQueryCheck := false
		dnsQueryLabel := widget.NewLabel("DNS Query:")
		dnsQueryEntry := widget.NewEntry()
		dnsQueryEntry.PlaceHolder = "Please input the DNS query domain name"
		dnsQueryCell := formCell(dnsQueryLabel, 100, dnsQueryEntry, 400)

		// dns protocol
		dnsProtocol := "udp"
		dnsProtocolLabel := widget.NewLabel("DNS Protocol:")
		dnsProtocolSelect := widget.NewSelect([]string{"udp", "tcp"}, func(s string) { dnsProtocol = s })
		dnsProtocolSelect.Selected = "udp"
		dnsProtocolCell := formCell(dnsProtocolLabel, 100, dnsProtocolSelect, 150)

		// specific container
		specificContainer.Add(dnsServerContainer)
		specificContainer.Add(dnsQueryCell)
		specificContainer.Add(dnsProtocolCell)

		// submit on Tap Action
		submitBtn.OnTapped = func() {

			// dns target validation
			dnsServers, err := targetHostValidator(dnsServerEntry.Text, true)
			if err != nil {
				dnsServerCheck = false
				errMsg.Text = err.Error()
				errMsg.Refresh()
				return
			}

			// valid IP check - no domain name is allowed
			for _, dnsServer := range dnsServers {
				errIn := IsValidIP(dnsServer)
				if errIn != nil {
					dnsServerCheck = false
					errMsg.Text = fmt.Sprintf("DNS server must be a valid IP address. Invalid input: %s", dnsServer)
					errMsg.Refresh()
					return
				}
			}

			dnsServerCheck = true
			errMsg.Text = ""
			errMsg.Refresh()

			// dns query validation
			if dnsQueryEntry.Text == "" {
				dnsQueryCheck = false
				errMsg.Text = "DNS Query cannot be empty!"
				errMsg.Refresh()
				return
			} else {
				dnsQueryCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// validation check
			if intervalCheck && timeoutCheck && dnsServerCheck && dnsQueryCheck {
				for _, dnsServer := range dnsServers {
					iv := ntPinger.InputVars{}
					iv.Type = "dns"
					iv.Count = 0
					iv.Dns_Protocol = dnsProtocol
					iv.Timeout = timeoutValue
					iv.Interval = intervalValue
					iv.DestHost = dnsServer
					iv.Dns_query = dnsQueryEntry.Text

					// start test
					go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, recording, db, entryChan, errChan)

					// close new test window
					newTestWindow.Close()
				}
			} else {
				return
			}
		}

	case "http":

		// update the New Test Window Size
		newTestWindow.Resize(fyne.NewSize(710, 300))

		// HTTP Method protocol
		httpMethod := "GET"
		httpMethodLabel := widget.NewLabel("HTTP Method:")
		httpMethodSelect := widget.NewSelect([]string{"GET", "PUT", "POST"}, func(s string) { httpMethod = s })
		httpMethodSelect.Selected = "GET"
		httpMethodCell := formCell(httpMethodLabel, 100, httpMethodSelect, 150)

		// URL
		httpURLCheck := false
		httpURLLabel := widget.NewLabel("Test URL:")

		httpURLNote01 := widget.NewLabel("Note: By default, HTTP uses TCP port 80, and HTTPS uses TCP port 443. If a different port is used, it must be specified after the domain name and before the path (if present).")
		httpURLNote01.Wrapping = fyne.TextWrapWord
		httpURLNote01.Resize(fyne.NewSize(300, 15))
		//httpURLNote01.TextStyle.Bold = true

		httpURLNote02 := widget.NewLabel("Example 1: www.google.com")
		httpURLNote02.Wrapping = fyne.TextWrapWord
		httpURLNote02.Resize(fyne.NewSize(300, 3))

		httpURLNote03 := widget.NewLabel("Example 2: www.mywebsite.com:8443/web/img")
		httpURLNote03.Wrapping = fyne.TextWrapWord
		httpURLNote03.Resize(fyne.NewSize(300, 3))

		httpURLNoteContainer := container.New(layout.NewVBoxLayout(), httpURLNote01, httpURLNote02, httpURLNote03)
		httpUrlNoteCard := widget.NewCard("", "", httpURLNoteContainer)

		httpScheme := "https://"
		httpSchemeSelect := widget.NewSelect([]string{"https://", "http://"}, func(s string) { httpScheme = s })
		httpSchemeSelect.Selected = "https://"
		httpURLEntry := widget.NewEntry()
		httpURLEntry.SetPlaceHolder("Please input the test URL")
		httpURLEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 40)), httpURLEntry)
		httpURLSchemeURLEntryContainer := container.New(layout.NewBorderLayout(nil, nil, httpSchemeSelect, nil), httpSchemeSelect, httpURLEntryContainer)
		httpURLContainer := container.New(layout.NewVBoxLayout(), httpURLLabel, httpURLSchemeURLEntryContainer, httpUrlNoteCard)

		// specific container
		specificContainer.Add(httpMethodCell)
		specificContainer.Add(httpURLContainer)

		// submit on Tap Action
		submitBtn.OnTapped = func() {
			// target URL validation
			httpInputVars, err := ParseTargetURL(fmt.Sprintf("%s%s", httpScheme, httpURLEntry.Text))
			if err != nil {
				httpURLCheck = false
				errMsg.Text = err.Error()
				errMsg.Refresh()
				return
			} else {
				httpURLCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// validation check
			if intervalCheck && timeoutCheck && httpURLCheck {

				// construct iput var
				iv := ntPinger.InputVars{}
				iv.Type = "http"
				iv.Count = 0
				iv.Timeout = timeoutValue
				iv.Interval = intervalValue
				iv.Http_method = httpMethod
				iv.DestHost = httpInputVars.Hostname
				iv.Http_path = httpInputVars.Path
				iv.Http_scheme = httpInputVars.Scheme
				iv.DestPort = httpInputVars.Port

				// start test
				go HttpAddPingRow(a, &ntGlobal.httpIndex, &iv, ntGlobal.httpTable, recording, db, entryChan, errChan)

				// close new test window
				newTestWindow.Close()

			} else {
				return
			}
		}
	case "tcp":
		// update the New Test Window Size
		newTestWindow.Resize(fyne.NewSize(710, 550))

		// target server
		tcpServerCheck := false
		tcpServerLabel := widget.NewLabel("TCP Server IP/Host(s):")
		tcpServerEntry := widget.NewMultiLineEntry()
		tcpServerEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 150)), tcpServerEntry)
		tcpServerContainer := container.New(layout.NewVBoxLayout(), tcpServerLabel, tcpServerEntryContainer)

		// tcp Port
		tcpPortCheck := false
		tcpPortLabel := widget.NewLabel("TCP Port:")
		tcpPortEntry := widget.NewEntry()
		tcpPortEntry.Resize(fyne.NewSize(100, 3))
		tcpPortCell := formCell(tcpPortLabel, 100, tcpPortEntry, 100)

		// specific container
		specificContainer.Add(tcpServerContainer)
		specificContainer.Add(tcpPortCell)

		// submit on Tap Action
		submitBtn.OnTapped = func() {

			// tcp Server validation
			tcpServers, err := targetHostValidator(tcpServerEntry.Text, true)
			if err != nil {
				tcpServerCheck = false
				errMsg.Text = err.Error()
				errMsg.Refresh()
				return
			} else {
				tcpServerCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// target Port Validation
			var tcpPort int
			if tcpPortEntry.Text == "" {
				tcpPortCheck = false
				errMsg.Text = "TCP Port cannot be empty!"
				errMsg.Refresh()
				return
			} else {
				tcpPort, err = strconv.Atoi(tcpPortEntry.Text)
				if err != nil {
					tcpPortCheck = false
					errMsg.Text = fmt.Sprintf("Invalid Port number: %s. Valid number is a integer", tcpPortEntry.Text)
					errMsg.Refresh()
					return
				}

				if tcpPort > 65535 || tcpPort < 1 {
					tcpPortCheck = false
					errMsg.Text = fmt.Sprintf("Invalid port number: %s. Valid range is 1–65535", tcpPortEntry.Text)
					errMsg.Refresh()
					return
				}

				tcpPortCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// validation check
			if tcpServerCheck && tcpPortCheck {
				for _, tcpServer := range tcpServers {
					iv := ntPinger.InputVars{}
					iv.Type = "tcp"
					iv.Count = 0
					iv.Timeout = timeoutValue
					iv.Interval = intervalValue
					iv.DestHost = tcpServer
					iv.DestPort = tcpPort
					iv.PayLoadSize = 0

					// start test
					go TcpAddPingRow(a, &ntGlobal.tcpIndex, &iv, ntGlobal.tcpTable, recording, db, entryChan, errChan)

					// close new test window
					newTestWindow.Close()
				}
			} else {
				return
			}
		}
	case "icmp":
		// update the New Test Window Size
		newTestWindow.Resize(fyne.NewSize(710, 550))

		// target server
		icmpServerCheck := false
		icmpServerLabel := widget.NewLabel("ICMP Server IP/Host(s):")
		icmpServerEntry := widget.NewMultiLineEntry()
		icmpServerEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 150)), icmpServerEntry)
		icmpServerContainer := container.New(layout.NewVBoxLayout(), icmpServerLabel, icmpServerEntryContainer)

		// icmp DF
		icmpDFLabel := widget.NewLabel("DF bit Set:")
		icmpDFBit := "OFF"
		icmpDFSelect := widget.NewSelect([]string{"OFF", "ON"}, func(s string) { icmpDFBit = s })
		icmpDFSelect.Selected = "OFF"
		icmpDFCell := formCell(icmpDFLabel, 100, icmpDFSelect, 100)

		// icmp Payload
		icmpPayloadCheck := false
		icmpPayloadLabel := widget.NewLabel("PayloadSize:")
		icmpPayloadEntry := widget.NewEntry()
		icmpPayloadEntry.Text = "32"
		icmpPayloadEntry.Resize(fyne.NewSize(100, 3))
		icmpPayloadCell := formCell(icmpPayloadLabel, 100, icmpPayloadEntry, 100)

		// specific container
		specificContainer.Add(icmpServerContainer)
		specificContainer.Add(icmpDFCell)
		specificContainer.Add(icmpPayloadCell)

		// submit on Tap Action
		submitBtn.OnTapped = func() {

			// icmp Server validation
			icmpServers, err := targetHostValidator(icmpServerEntry.Text, true)
			if err != nil {
				icmpServerCheck = false
				errMsg.Text = err.Error()
				errMsg.Refresh()
				return
			} else {
				icmpServerCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// tcp payload validation
			var icmpPayloadSize int
			icmpPayloadSize, err = strconv.Atoi(icmpPayloadEntry.Text)
			if err != nil {
				icmpPayloadCheck = false
				errMsg.Text = fmt.Sprintf("Invalid Payload Size: %s. Valid number is a integer", icmpPayloadEntry.Text)
				errMsg.Refresh()
				return
			}

			if icmpPayloadSize > 65535 || icmpPayloadSize < 32 {
				icmpPayloadCheck = false
				errMsg.Text = fmt.Sprintf("Invalid Payload Size: %s. Valid range is 32–65535", icmpPayloadEntry.Text)
				errMsg.Refresh()
				return
			} else {
				icmpPayloadCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

			// validation check
			if icmpServerCheck && icmpPayloadCheck {

				// set DF Bit
				var DFBit bool
				if icmpDFBit == "OFF" {
					DFBit = false
				} else {
					DFBit = true
				}

				// create ping row
				for _, icmpServer := range icmpServers {
					iv := ntPinger.InputVars{}
					iv.Type = "icmp"
					iv.Count = 0
					iv.Timeout = timeoutValue
					iv.Interval = intervalValue
					iv.DestHost = icmpServer
					iv.Icmp_DF = DFBit
					iv.PayLoadSize = icmpPayloadSize

					// start test
					go IcmpAddPingRow(a, &ntGlobal.icmpIndex, &iv, ntGlobal.icmpTable, recording, db, entryChan, errChan)

					// close new test window
					newTestWindow.Close()
				}
			} else {
				return
			}
		}

	}

	// New Test Input Container
	newTestSpaceHolder := widget.NewLabel("                     ")
	newTestContainerInnerUp := container.New(layout.NewVBoxLayout(), commonContainer, specificContainer, errMsgContainer)
	newTestContainerInnerdown := container.NewCenter(btnContainer)
	newTestContainerInnerWhole := container.New(layout.NewVBoxLayout(), newTestContainerInnerUp, newTestContainerInnerdown)
	newTestContainerOuter := container.New(layout.NewBorderLayout(newTestSpaceHolder, newTestSpaceHolder, newTestSpaceHolder, newTestSpaceHolder), newTestSpaceHolder, newTestContainerInnerWhole)
	newTestWindow.SetContent(newTestContainerOuter)
	newTestWindow.Show()
}
