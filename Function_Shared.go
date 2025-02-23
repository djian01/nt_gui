package main

import (
	"fmt"
	"image/color"
	"net"
	"net/url"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
