package main

import (
	"fmt"
	"image/color"
	"net"
	"net/url"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
)

// Func: Create a vertical separator
func GUIVerticalSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.RGBA{200, 200, 200, 255}) // Gray color for a subtle look
	separator.SetMinSize(fyne.NewSize(1, 30))                        // 2px width, full height of row
	return separator
}

// Func: Generate NT CMD
func NtCmdGenerator(recording bool, iv ntPinger.InputVars) string {

	// initial Cmd
	ntCmd := ""
	CmdRecording := ""
	CmdInterval := ""
	CmdTimeout := ""

	// recording

	if recording {
		CmdRecording = " -r"
	}

	// switch based on Type
	switch iv.Type {
	case "dns":
		// Interval
		if iv.Interval != 1 {
			CmdInterval = fmt.Sprintf(" -i %v", iv.Interval)
		}
		// Timeout
		if iv.Timeout != 4 {
			CmdTimeout = fmt.Sprintf(" -t %v", iv.Timeout)
		}
		// DnsProtocol
		CmdDnsProtocol := ""
		if iv.Dns_Protocol == "tcp" {
			CmdDnsProtocol = " -o tcp"
		}
		// ntCmd
		ntCmd = fmt.Sprintf("nt%s dns%s%s%s %s %s", CmdRecording, CmdDnsProtocol, CmdInterval, CmdTimeout, iv.DestHost, iv.Dns_query)

	case "http":
		// Interval
		if iv.Interval != 5 {
			CmdInterval = fmt.Sprintf(" -i %v", iv.Interval)
		}
		// Timeout
		if iv.Timeout != 4 {
			CmdTimeout = fmt.Sprintf(" -t %v", iv.Timeout)
		}
		// HttpMethod
		CmdHttpMethod := ""
		if iv.Http_method != "GET" {
			CmdHttpMethod = fmt.Sprintf(" -m %s", iv.Http_method)
		}

		// Http URL
		httpUrl := ntPinger.ConstructURL(iv.Http_scheme, iv.DestHost, iv.Http_path, iv.DestPort)

		// ntCmd
		ntCmd = fmt.Sprintf("nt%s http%s%s%s %s %s", CmdRecording, CmdHttpMethod, CmdInterval, CmdTimeout, iv.DestHost, httpUrl)

	case "tcp":
		// Interval
		if iv.Interval != 1 {
			CmdInterval = fmt.Sprintf(" -i %v", iv.Interval)
		}
		// Timeout
		if iv.Timeout != 4 {
			CmdTimeout = fmt.Sprintf(" -t %v", iv.Timeout)
		}

		// ntCmd
		ntCmd = fmt.Sprintf("nt%s tcp%s%s %s %v", CmdRecording, CmdInterval, CmdTimeout, iv.DestHost, iv.DestPort)

	case "icmp":
		// Interval
		if iv.Interval != 1 {
			CmdInterval = fmt.Sprintf(" -i %v", iv.Interval)
		}
		// Timeout
		if iv.Timeout != 4 {
			CmdTimeout = fmt.Sprintf(" -t %v", iv.Timeout)
		}
		// ICMP Payload
		IcmpPayloadSize := ""
		if iv.PayLoadSize != 32 {
			IcmpPayloadSize = fmt.Sprintf(" -s %v", iv.PayLoadSize)
		}
		// ICMP DF
		IcmpDf := ""
		if iv.Icmp_DF {
			IcmpDf = " -d"
		}

		// ntCmd
		ntCmd = fmt.Sprintf("nt%s icmp%s%s%s%s %s", CmdRecording, IcmpDf, IcmpPayloadSize, CmdInterval, CmdTimeout, iv.DestHost)
	}

	return ntCmd
}

// TruncateString truncates a string to a maximum length and appends "..." if it exceeds the max length
func TruncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength-3] + "..." // Subtract 3 to account for "..."
	}
	return s
}

// func Prsae URL
func parseURL(urlStr string) *url.URL {
	link, _ := url.Parse(urlStr)
	return link
}

// func: New Test Input
func NewTest(a fyne.App, testType string, testTable *fyne.Container) {

	// Initial New Test Input Var Window
	newTestWindow := a.NewWindow(fmt.Sprintf("New %s Test", testType))
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
	intervalLabel := widget.NewLabel("Interval (s):")
	intervalEntry := widget.NewEntry()
	intervalEntry.Text = "1"
	intervalEntry.Validator = func(s string) error {
		// Convert text to an integer
		num, err := strconv.Atoi(s)
		if err != nil || num < 1 {
			msg := "interval should always be Int and larger than 0"
			errMsg.Text = msg
			errMsg.Refresh()
			return fmt.Errorf("validation error: %s", msg)
		} else {
			errMsg.Text = ""
			errMsg.Refresh()
			return nil
		}
	}

	intervalContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), intervalEntry)
	intervalCell := formCell(intervalLabel, 100, intervalContainer, 100)

	// Timeout
	timeoutLabel := widget.NewLabel("Timeout (s):")
	timeoutEntry := widget.NewEntry()
	timeoutEntry.Text = "4"
	timeoutEntry.Validator = func(s string) error {
		// Convert text to an integer
		num, err := strconv.Atoi(s)
		if err != nil || num < 1 {
			msg := "Timeout should always be Int and larger than 0"
			errMsg.Text = msg
			errMsg.Refresh()
			return fmt.Errorf("validation error: %s", msg)
		} else {
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
		if b {
			recording = true
		} else {
			recording = false
		}
		fmt.Println(recording)
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
		dnsServerLabel := widget.NewLabel("DNS Server IP/Host(s):")
		dnsServerEntry := widget.NewMultiLineEntry()
		dnsServerEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(504, 150)), dnsServerEntry)
		dnsServerContainer := container.New(layout.NewVBoxLayout(), dnsServerLabel, dnsServerEntryContainer)

		// dns query
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
			dnsServers, err := targetHostValidator(dnsServerEntry.Text, true)
			if err != nil {
				errMsg.Text = err.Error()
				errMsg.Refresh()
				return
			} else {
				errMsg.Text = ""
				errMsg.Refresh()
			}

			for _, i := range dnsServers {
				fmt.Println(i)
			}
			fmt.Println(dnsProtocol)
		}

	case "http":
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

// func: create a 2 x Column form cell
func formCell(obj1 fyne.CanvasObject, length1 float32, obj2 fyne.CanvasObject, length2 float32) *fyne.Container {
	// object 1
	obj1Container := container.New(layout.NewGridWrapLayout(fyne.NewSize(length1, 40)), obj1)

	// object 2
	obj2Container := container.New(layout.NewGridWrapLayout(fyne.NewSize(length2, 40)), obj2)

	// form Cell Container

	formCellContainer := container.New(layout.NewHBoxLayout(), obj1Container, obj2Container)
	return formCellContainer
}

// func: ValidateAndResolve checks if the input string is a valid IP or a resolvable DNS name
func ValidateAndResolve(input string, requiredResolve bool) (string, error) {
	// Step 1: Check if the string is a valid IP
	if ip := net.ParseIP(input); ip != nil {
		return input, nil // Valid IP, return as is
	}

	// Step 2: Check if the string is a resolvable DNS name
	ips, err := net.LookupIP(input)
	if err != nil {
		return "", fmt.Errorf("Bad IP or unresolvable Domain: %v", input) // Error Message
	}

	// Step 3: Return based on requiredResolve flag
	if requiredResolve {
		return ips[0].String(), nil // Return the first resolved IP
	} else {
		return input, nil // Return original string if no resolution is required
	}

}

// func: dns Server Input Validator
func targetHostValidator(inputTargets string, requiredResolve bool) (targetHosts []string, err error) {

	if inputTargets == "" {
		return targetHosts, fmt.Errorf("No Input IP/Host Target")
	}

	targetsTemp := regexp.MustCompile(`\r?\n`).Split(inputTargets, -1)

	for _, input := range targetsTemp {
		server, err := ValidateAndResolve(input, requiredResolve)
		if err != nil {
			return targetHosts, err
		} else {
			targetHosts = append(targetHosts, server)
		}
	}
	return targetHosts, nil
}

// func: create a place holder
func placeHolderBlock(w, h float32) *canvas.Rectangle {
	placeholder := canvas.NewRectangle(theme.Color(theme.ColorNameBackground)) // Matches background color
	placeholder.SetMinSize(fyne.NewSize(w, h))                                 // Set fixed size

	return placeholder
}
