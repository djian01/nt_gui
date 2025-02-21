package main

import (
	"fmt"
	"image/color"
	"net/url"
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
	newTestWindow.Resize(fyne.NewSize(1200, 650))

	// error message
	errMsg := canvas.NewText("", color.RGBA{255, 0, 0, 255})
	errMsg.TextStyle.Bold = true
	errMsg.Text = ""
	errMsgContainer := container.New(layout.NewVBoxLayout(), errMsg)

	// common container (common for all test types)
	// Interval
	intervalLabel := widget.NewLabel("Interval (s)")
	intervalEntry := widget.NewEntry()
	intervalEntry.Text = "1"
	intervalEntry.OnChanged = func(text string) {
		// Convert text to an integer
		num, err := strconv.Atoi(intervalEntry.Text)
		if err != nil || num < 1 {
			errMsg.Text = "Interval should always be Int and larger than 0"
			errMsg.Refresh()
		} else {
			errMsg.Text = ""
			errMsg.Refresh()
		}
	}
	intervalContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), intervalEntry)
	intervalCell := formCell(intervalLabel, intervalContainer)

	// Timeout
	timeoutLabel := widget.NewLabel("Timeout (s)")
	timeoutEntry := widget.NewEntry()
	timeoutEntry.Text = "4"
	timeoutEntry.OnChanged = func(text string) {
		// Convert text to an integer
		num, err := strconv.Atoi(timeoutEntry.Text)
		if err != nil || num < 1 {
			errMsg.Text = "Timeout should always be Int and larger than 0"
			errMsg.Refresh()
		} else {
			errMsg.Text = ""
			errMsg.Refresh()
		}
	}
	timeoutContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), timeoutEntry)
	timeoutCell := formCell(timeoutLabel, timeoutContainer)
	commonContainer := container.NewHBox(intervalCell, timeoutCell)

	// Specific Vars
	specificContainer := container.NewHBox()

	// btns
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {})
	cancelBtn.Importance = widget.WarningImportance
	submitBtn := widget.NewButtonWithIcon("Submit", theme.ConfirmIcon(), func() {})
	submitBtn.Importance = widget.HighImportance
	btnContainer := formCell(cancelBtn, submitBtn)

	// New Test Input Container
	newTestSpaceHolder := widget.NewLabel("                     ")
	newTestContainerInner := container.New(layout.NewVBoxLayout(), commonContainer, specificContainer, btnContainer, errMsgContainer)
	newTestContainerOuter := container.New(layout.NewBorderLayout(newTestSpaceHolder, newTestSpaceHolder, newTestSpaceHolder, newTestSpaceHolder), newTestSpaceHolder, newTestContainerInner)
	newTestWindow.SetContent(newTestContainerOuter)
	newTestWindow.Show()

}

// func: create a 2 x Column form cell
func formCell(obj1, obj2 fyne.CanvasObject) *fyne.Container {
	//formCellContainer := container.NewCenter(container.NewGridWrap(fyne.NewSize(length, 30), container.New(layout.NewHBoxLayout(), obj1, obj2)))
	formCellContainer := container.NewCenter(container.New(layout.NewHBoxLayout(), obj1, obj2))
	return formCellContainer
}
