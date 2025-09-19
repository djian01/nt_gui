package main

import (
	"crypto/rand"
	"fmt"
	"image/color"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"github.com/djian01/nt/pkg/ntPinger"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	"github.com/kbinani/screenshot"
)

// Func: get the config file path for different OS
func getConfigFilePath(appName string) (string, error) {
	var configDir string
	var err error

	if runtime.GOOS == "darwin" {
		// macOS: ~/Library/Application Support/<appName> (/Users/<User Name>/Library/Application Support/<appName>)
		configDir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(configDir, appName)
	} else {
		// Windows/Linux: directory where executable resides
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}
		configDir = filepath.Dir(exePath)
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return "", err
	}

	// Return full path for config file
	return configDir, nil
}

// Func: Create a vertical separator
func GUIVerticalSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.RGBA{200, 200, 200, 255}) // Gray color for a subtle look
	separator.SetMinSize(fyne.NewSize(1, 30))                        // 2px width, full height of row
	return separator
}

// Func: Generate NT CMD by InputVars
func Iv2NtCmd(recording bool, iv ntPinger.InputVars) string {

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

		// HttpStatusCode
		CmdHttpStatusCode := ""
		if len(iv.Http_statusCodes) > 0 {
			for _, s := range iv.Http_statusCodes {
				CmdHttpStatusCode += fmt.Sprintf(" -s %s", s.Code2String())
			}
		}

		// Http URL
		httpUrl := ntPinger.ConstructURL(iv.Http_scheme, iv.DestHost, iv.Http_path, iv.DestPort)

		// ntCmd
		ntCmd = fmt.Sprintf("nt%s http%s%s%s%s %s", CmdRecording, CmdHttpMethod, CmdHttpStatusCode, CmdInterval, CmdTimeout, httpUrl)

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

// Func: Generate InputVars from NT CMD
func NtCmd2Iv(cmd string) (recording bool, iv ntPinger.InputVars, err error) {

	// Split command into parts based on the white spaces
	parts := strings.Fields(cmd)
	if len(parts) < 2 || parts[0] != "nt" {
		return false, iv, fmt.Errorf("invalid command format")
	}

	// Check for recording flag
	if parts[1] == "-r" {
		recording = true
		parts = append([]string{}, parts[2:]...) // Remove "nt -r"
	} else if parts[0] == "nt" {
		parts = parts[1:] // Remove "nt"
	} else {
		return false, iv, fmt.Errorf("invalid command format, should start with 'nt'")
	}

	// Command type
	if len(parts) < 1 {
		return false, iv, fmt.Errorf("missing command type")
	}
	iv.Type = parts[0]
	args := parts[1:]

	// Iterate over arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c":
			if i+1 < len(args) {
				iv.Count, _ = strconv.Atoi(args[i+1])
			}
		case "-i":
			if i+1 < len(args) {
				iv.Interval, _ = strconv.Atoi(args[i+1])
			}
		case "-t":
			if i+1 < len(args) {
				iv.Timeout, _ = strconv.Atoi(args[i+1])
			}
		case "-o":
			if i+1 < len(args) && iv.Type == "dns" {
				iv.Dns_Protocol = args[i+1]
			}
		case "-m":
			if i+1 < len(args) && iv.Type == "http" {
				iv.Http_method = args[i+1]
			}
		case "-s":
			if i+1 < len(args) && iv.Type == "icmp" {
				iv.PayLoadSize, _ = strconv.Atoi(args[i+1])
			}
		case "-d":
			if iv.Type == "icmp" {
				iv.Icmp_DF = true
			}
		default:
			// Assign destination and additional parameters

			if iv.Type == "icmp" {
				iv.DestHost = args[len(args)-1]
			} else if iv.Type == "tcp" {
				iv.DestHost = args[len(args)-2]
				iv.DestPort, _ = strconv.Atoi(args[len(args)-1])
			} else if iv.Type == "http" {
				httpVars, _ := ParseURL2HttpVars(args[len(args)-1])
				iv.DestHost = httpVars.Hostname
				iv.Http_path = httpVars.Path
				iv.Http_scheme = httpVars.Scheme
				iv.DestPort = httpVars.Port

			} else if iv.Type == "dns" {
				iv.DestHost = args[len(args)-2]
				iv.Dns_query = args[len(args)-1]
				if iv.Dns_Protocol == "" {
					iv.Dns_Protocol = "udp"
				}
			}
		}
	}

	// Set default values if not specified
	if iv.Interval == 0 {
		switch iv.Type {
		case "icmp", "dns", "tcp":
			iv.Interval = 1
		case "http":
			iv.Interval = 5
		}
	}
	if iv.Timeout == 0 {
		iv.Timeout = 4
	}
	if iv.Http_method == "" && iv.Type == "http" {
		iv.Http_method = "GET"
	}
	if iv.PayLoadSize == 0 && iv.Type == "icmp" {
		iv.PayLoadSize = 32
	}

	return recording, iv, nil
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
		// ignore any whitespace line(s)
		if input == "" {
			continue
		}

		// drop whitespace inside the string
		input = strings.TrimSpace(input)

		// Validate and Resolve
		_, err := ValidateAndResolve(input, requiredResolve)
		if err != nil {
			return targetHosts, err
		} else {
			targetHosts = append(targetHosts, input)
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

// ParseURL extracts scheme, hostname, port, and path from a URL and build a new HttpVars
func ParseURL2HttpVars(inputURL string) (HttpVars, error) {

	HttpVarNew := HttpVars{}

	parsedURL, err := url.Parse(inputURL)

	if err != nil {
		return HttpVarNew, err
	}

	HttpVarNew.Scheme = parsedURL.Scheme
	HttpVarNew.Hostname = parsedURL.Hostname()

	// Handle default ports for http and https
	if parsedURL.Port() != "" {
		HttpVarNew.Port, err = strconv.Atoi(parsedURL.Port())
		if err != nil {
			return HttpVarNew, err
		}
	} else if HttpVarNew.Scheme == "http" {
		HttpVarNew.Port = 80
	} else if HttpVarNew.Scheme == "https" {
		HttpVarNew.Port = 443
	}

	if parsedURL.Path != "" {
		HttpVarNew.Path = parsedURL.Path
	}

	return HttpVarNew, nil
}

// func: slice clone for []ntchart.ChartPoint
func CloneChartPoints(chartPoints *[]ntchart.ChartPoint) []ntchart.ChartPoint {

	// create a new snapshot slice with fixed length
	chartPointSnapshot := make([]ntchart.ChartPoint, len(*chartPoints))

	// copy the chartPoints to snapshot
	copy(chartPointSnapshot, *chartPoints)

	return chartPointSnapshot
}

// Test Register Func: Check if the given UUID exists in the Test Register
func existingTestCheck(testRegister *[]string, uuid string) bool {
	for _, str := range *testRegister {
		if str == uuid {
			return true
		}
	}
	return false
}

// Test Register Func: delete UUID from a Test Register
func UnregisterTest(testRegister *[]string, uuid string) {
	newSlice := (*testRegister)[:0] // Keep the same underlying array
	for _, str := range *testRegister {
		if str != uuid {
			newSlice = append(newSlice, str)
		}
	}
	*testRegister = newSlice // Update the original slice
}

// GenerateShortUUID generates a 6-character alphanumeric (Base-62) UUID
func GenerateShortUUID() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 6)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

// ParseURL extracts scheme, hostname, port, and path from a URL
func ParseURL(inputURL string) (HttpVars, error) {

	HttpVarNew := HttpVars{}

	parsedURL, err := url.Parse(inputURL)

	if err != nil {
		return HttpVarNew, err
	}

	HttpVarNew.Scheme = parsedURL.Scheme
	HttpVarNew.Hostname = parsedURL.Hostname()

	// Handle default ports for http and https
	if parsedURL.Port() != "" {
		HttpVarNew.Port, err = strconv.Atoi(parsedURL.Port())
		if err != nil {
			return HttpVarNew, err
		}
	} else if HttpVarNew.Scheme == "http" {
		HttpVarNew.Port = 80
	} else if HttpVarNew.Scheme == "https" {
		HttpVarNew.Port = 443
	}

	if parsedURL.Path != "" {
		HttpVarNew.Path = parsedURL.Path
	}

	return HttpVarNew, nil
}

// func: ParseTargetURL (Parse Target Test URL and return errors if required)
func ParseTargetURL(inputURL string) (HttpVars, error) {
	testHttpVar, err := ParseURL(inputURL)
	if err != nil {
		return testHttpVar, err
	}

	// http scheme check
	schemeCheck := false
	if testHttpVar.Scheme == "http" || testHttpVar.Scheme == "https" {
		schemeCheck = true
	} else {
		return testHttpVar, fmt.Errorf("Invalid http scheme. It has to be either http or https!")
	}

	// http host check
	hostCheck := false
	_, err = net.LookupHost(testHttpVar.Hostname)
	if err != nil {
		return testHttpVar, err
	} else {
		hostCheck = true
	}

	// return
	if schemeCheck && hostCheck {
		return testHttpVar, nil
	} else {
		return testHttpVar, fmt.Errorf("Invalid input url.")
	}
}

// func: valid IP address check
func IsValidIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return nil
}

// function: set the max window size
func getPrimaryScreenSize() (fyne.Size, error) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return fyne.Size{}, fmt.Errorf("no active displays found")
	}

	bounds := screenshot.GetDisplayBounds(0) // Primary display

	return fyne.NewSize(float32(bounds.Dx())*0.4, float32(bounds.Dy())*0.4), nil
}
