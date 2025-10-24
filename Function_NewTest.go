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
	"github.com/djian01/nt_gui/pkg/ntwidget"
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
	recordingLabel := widget.NewLabel("Result Recording OFF")
	recordingSwitch := ntwidget.NewToggleswitch(false, func(b bool) {
		recording = b
		if b {
			fyne.Do(func() {
				recordingLabel.SetText("Result Recording ON")
			})
		} else {
			fyne.Do(func() {
				recordingLabel.SetText("Result Recording OFF")
			})
		}
	})

	recordingRow := ntwidget.ToggleRow(recordingSwitch, recordingLabel, 520) // set rowW to match your form width

	// common Container
	commonContainerSub := formCell(intervalCell, 250, timeoutCell, 250)
	commonContainer := container.NewVBox(recordingRow, commonContainerSub)

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
		newTestWindow.Resize(fyne.NewSize(710, 840))

		// HTTP Proxy
		// recording
		httpProxyFlag := false
		httpProxyStr := "none"

		//// HTTP Proxy Form
		// create entries
		proxyServerEntry := widget.NewEntry()
		proxyServerEntry.SetPlaceHolder("Example: http://proxy.example.com")

		proxyPortEntry := widget.NewEntry()
		proxyPortEntry.SetPlaceHolder("Example: 8080")

		proxyUserEntry := widget.NewEntry()
		proxyUserEntry.SetPlaceHolder("Optional")

		proxyPassEntry := widget.NewPasswordEntry()
		proxyPassEntry.SetPlaceHolder("Optional")

		// create labeled rows
		proxyServerRow := formCell(widget.NewLabel("Proxy Server:"), 130, proxyServerEntry, 450)
		proxyPortRow := formCell(widget.NewLabel("Proxy Port:"), 130, proxyPortEntry, 250)
		proxyUserRow := formCell(widget.NewLabel("Proxy Username:"), 130, proxyUserEntry, 250)
		proxyPassRow := formCell(widget.NewLabel("Proxy Password:"), 130, proxyPassEntry, 250)

		// group them together
		proxyForm := container.NewVBox(
			proxyServerRow,
			proxyPortRow,
			//proxyServerPortRow,
			proxyUserRow,
			proxyPassRow,
		)

		proxyForm.Hidden = true

		//// HTTP Proxy Switch

		httpProxyLabel := widget.NewLabel("Use HTTP Proxy: OFF")
		httpProxySwitch := ntwidget.NewToggleswitch(false, func(b bool) {
			httpProxyFlag = b
			if b {
				fyne.Do(func() {
					proxyForm.Hidden = false
					httpProxyLabel.SetText("Use HTTP Proxy: ON")
				})
			} else {
				fyne.Do(func() {
					proxyForm.Hidden = true
					httpProxyLabel.SetText("Use HTTP Proxy: OFF")
				})
			}
		})

		httpProxySwitchRow := ntwidget.ToggleRow(httpProxySwitch, httpProxyLabel, 520) // set rowW to match your form width

		httpProxyContainer := container.NewVBox(httpProxySwitchRow, proxyForm)
		httpProxyCard := widget.NewCard("", "", httpProxyContainer)

		specificContainer.Add(httpProxyCard)

		// HTTP Status Codes
		httpStatusLabel := widget.NewLabel("Expected HTTP Status Codes:")

		s2xxLabel := widget.NewLabel("2xx")
		s2xxCheck := widget.NewCheck("", func(b bool) {})
		s2xxCheck.SetChecked(true)
		s2xxCell := formCell(s2xxLabel, 30, s2xxCheck, 80)

		s3xxLabel := widget.NewLabel("3xx")
		s3xxCheck := widget.NewCheck("", func(b bool) {})
		s3xxCheck.SetChecked(true)
		s3xxCell := formCell(s3xxLabel, 30, s3xxCheck, 80)

		s4xxLabel := widget.NewLabel("4xx")
		s4xxCheck := widget.NewCheck("", func(b bool) {})
		s4xxCell := formCell(s4xxLabel, 30, s4xxCheck, 80)

		s5xxLabel := widget.NewLabel("5xx")
		s5xxCheck := widget.NewCheck("", func(b bool) {})
		s5xxCell := formCell(s5xxLabel, 30, s5xxCheck, 80)

		sCustomLabel := widget.NewLabel("Custom (Optional)")
		sCustomCheck := true
		sCustomEntry := widget.NewEntry()
		sCustomEntry.SetPlaceHolder("Enter a value between 200 and 599")
		sCustomEntry.OnChanged = func(s string) {
			// Custom Status Code Entry is Empty
			if s == "" {
				sCustomCheck = true // allow clearing
				errMsg.Text = ""
				errMsg.Refresh()
				return
			}
			// Custom Status Code Entry is outside of the valid range
			if val, err := strconv.Atoi(s); err != nil || val < 200 || val > 599 {
				sCustomCheck = false
				errMsg.Text = "HTTP status code must be between 200 and 599."
				errMsg.Refresh()
				return
			}

			sCustomCheck = true
			errMsg.Text = ""
			errMsg.Refresh()
		}

		sCustomCell := formCell(sCustomLabel, 150, sCustomEntry, 260)

		httpStatusInUpContainer := container.New(layout.NewHBoxLayout(), s2xxCell, s3xxCell, s4xxCell, s5xxCell)
		httpStatusInDownContainer := container.New(layout.NewVBoxLayout(), httpStatusInUpContainer, sCustomCell)
		httpStatusOutContainer := container.New(layout.NewVBoxLayout(), httpStatusLabel, httpStatusInDownContainer)

		httpStatusCard := widget.NewCard("", "", httpStatusOutContainer)

		specificContainer.Add(httpStatusCard)

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

			// Custom Status Code Check
			statusCodeCheck := false
			httpStatusCodes := []ntPinger.HttpStatusCode{}

			if sCustomEntry.Text != "" {
				if val, err := strconv.Atoi(sCustomEntry.Text); err != nil {
					sCustomCheck = false
					errMsg.Text = err.Error()
					errMsg.Refresh()
					return
				} else {
					httpStatusCodes = append(httpStatusCodes, ntPinger.HttpStatusCode{LowerCode: val, UpperCode: val})
					statusCodeCheck = true
				}
			}

			// HTTP Proxy Check and httpProxyStr creation
			httpProxycheck := false

			if httpProxyFlag {
				httpProxyStr, err = BuildProxyURL(proxyServerEntry.Text, proxyPortEntry.Text, proxyUserEntry.Text, proxyPassEntry.Text)
				if err != nil {
					errMsg.Text = err.Error()
					fyne.Do(func() {
						errMsg.Refresh()
					})
					return
				} else {
					httpProxycheck = true
				}

			} else {
				httpProxycheck = true
				httpProxyStr = "none"
			}

			// Status Code Check
			if s2xxCheck.Checked {
				httpStatusCodes = append(httpStatusCodes, ntPinger.HttpStatusCode{LowerCode: 200, UpperCode: 299})
				statusCodeCheck = true
			}
			if s3xxCheck.Checked {
				httpStatusCodes = append(httpStatusCodes, ntPinger.HttpStatusCode{LowerCode: 300, UpperCode: 399})
				statusCodeCheck = true
			}
			if s4xxCheck.Checked {
				httpStatusCodes = append(httpStatusCodes, ntPinger.HttpStatusCode{LowerCode: 400, UpperCode: 499})
				statusCodeCheck = true
			}
			if s5xxCheck.Checked {
				httpStatusCodes = append(httpStatusCodes, ntPinger.HttpStatusCode{LowerCode: 500, UpperCode: 599})
				statusCodeCheck = true
			}

			if !statusCodeCheck {
				errMsg.Text = "Please Select at least one HTTP Status Code!"
				fyne.Do(func() {
					errMsg.Refresh()
				})
				return
			}
			// validation check
			if intervalCheck && timeoutCheck && httpURLCheck && sCustomCheck && statusCodeCheck && httpProxycheck {

				// construct iput var
				iv := ntPinger.InputVars{}
				iv.Type = "http"
				iv.Count = 0
				iv.Timeout = timeoutValue
				iv.Interval = intervalValue
				iv.Http_method = httpMethod
				iv.Http_statusCodes = httpStatusCodes
				iv.DestHost = httpInputVars.Hostname
				iv.Http_path = httpInputVars.Path
				iv.Http_scheme = httpInputVars.Scheme
				iv.DestPort = httpInputVars.Port
				iv.Http_proxy = httpProxyStr

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
