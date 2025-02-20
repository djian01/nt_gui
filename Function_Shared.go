package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
