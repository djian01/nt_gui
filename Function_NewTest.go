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
	newTestWindow.Resize(fyne.NewSize(710, 550))
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
			} else {
				dnsServerCheck = true
				errMsg.Text = ""
				errMsg.Refresh()
			}

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
		// HTTP Method protocol
		httpMethod := "GET"
		httpMethodLabel := widget.NewLabel("HTTP Method:")
		httpMethodSelect := widget.NewSelect([]string{"GET", "PUT", "POST"}, func(s string) { httpMethod = s })
		httpMethodSelect.Selected = "GET"
		httpMethodCell := formCell(httpMethodLabel, 100, httpMethodSelect, 150)

		// URL
		httpURLCheck := false
		httpURLLabel := widget.NewLabel("Test URL:")
		httpURLEntry := widget.NewEntry()
		httpURLEntry.SetPlaceHolder("Exapmle: https://www.mywebsite.com:8443/web/img")
		httpURLEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 150)), httpURLEntry)
		httpURLContainer := container.New(layout.NewVBoxLayout(), httpURLLabel, httpURLEntryContainer)

		// specific container
		specificContainer.Add(httpMethodCell)
		specificContainer.Add(httpURLContainer)

		// submit on Tap Action
		submitBtn.OnTapped = func() {
			// target URL validation
			httpInputVars, err := ParseTargetURL(httpURLEntry.Text)
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
				go HttpAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.httpTable, recording, db, entryChan, errChan)

				// close new test window
				newTestWindow.Close()

			} else {
				return
			}
		}
	case "tcp":
	case "icmp":

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
