package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

func SaveToCSV(filePath string, iv ntPinger.InputVars, dbTestEntries *[]ntdb.DbTestEntry) error {

	// Open or create the file with append mode and write-only access
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// if accumulatedRecords is empty
	if len(*dbTestEntries) == 0 {
		return nil
		// else Save to CSV based on Type
	} else {
		switch iv.Type {
		case "icmp":
			// Write the header
			header := []string{
				"Type",
				"Seq",
				"Status",
				"DestHost",
				"DestAddr",
				"PayLoadSize",
				"RTT(ms)",
				"SendDate",
				"SendTime",
				"PacketsSent",
				"PacketsRecv",
				"PacketLoss",
				"MinRtt",
				"AvgRtt",
				"MaxRtt",
				"AdditionalInfo",
			}
			err := writer.Write(header)
			if err != nil {
				return fmt.Errorf("could not write header to file: %v", err)
			}

			// Write each struct to the file
			for _, recordItem := range *dbTestEntries {
				// interface assertion
				pkt := recordItem.(*ntdb.RecordICMPEntry)

				// construct the row
				row := []string{
					iv.Type,                       // Ping Type
					strconv.Itoa(pkt.Seq),         // Seq
					fmt.Sprintf("%v", pkt.Status), // Status
					iv.DestHost,                   // DestHost
					iv.DestHost,                   // DestAddr
					strconv.Itoa(iv.PayLoadSize),  // PayLoadSize

					fmt.Sprintf("%v", float64(pkt.RTT.Nanoseconds())/1e6), // Response_Time
					pkt.GetSendTime().Format("2006-01-02"),                // SendDate
					pkt.GetSendTime().Format("15:04:05 MST"),              // SendTime

					strconv.Itoa(pkt.Seq + 1),    // PacketsSent
					strconv.Itoa(pkt.PacketRecv), // PacketsRecv
					pkt.PacketLossRate,           // PacketLoss
					pkt.MinRTT,                   // MinRtt
					pkt.AvgRTT,                   // AvgRtt
					pkt.MaxRTT,                   // MaxRtt
					pkt.AddInfo,                  // AdditionalInfo
				}

				if err := writer.Write(row); err != nil {
					return fmt.Errorf("could not write record to file: %v", err)
				}
			}
		case "tcp":
			// Write the header
			header := []string{
				"Type",
				"Seq",
				"Status",
				"DestHost",
				"DestAddr",
				"DestPort",
				"PayLoadSize",
				"RTT(ms)",
				"SendDate",
				"SendTime",
				"PacketsSent",
				"PacketsRecv",
				"PacketLoss",
				"MinRtt",
				"AvgRtt",
				"MaxRtt",
				"AdditionalInfo",
			}

			err := writer.Write(header)

			if err != nil {
				return fmt.Errorf("could not write header to file: %v", err)
			}

			// Write each struct to the file
			for _, recordItem := range *dbTestEntries {

				// interface assertion
				pkt := recordItem.(*ntdb.RecordTCPEntry)

				// construct the row
				row := []string{
					iv.Type,                       // Ping Type
					strconv.Itoa(pkt.Seq),         // Seq
					fmt.Sprintf("%v", pkt.Status), // Status
					iv.DestHost,                   // DestHost
					iv.DestHost,                   // DestAddr
					strconv.Itoa(iv.DestPort),     // DestPort
					strconv.Itoa(iv.PayLoadSize),  // PayLoadSize

					fmt.Sprintf("%v", float64(pkt.RTT.Nanoseconds())/1e6), // Response_Time
					pkt.GetSendTime().Format("2006-01-02"),                // SendDate
					pkt.GetSendTime().Format("15:04:05 MST"),              // SendTime

					strconv.Itoa(pkt.Seq + 1),    // PacketsSent
					strconv.Itoa(pkt.PacketRecv), // PacketsRecv
					pkt.PacketLossRate,           // PacketLoss
					pkt.MinRTT,                   // MinRtt
					pkt.AvgRTT,                   // AvgRtt
					pkt.MaxRTT,                   // MaxRtt
					pkt.AddInfo,                  // AdditionalInfo
				}

				if err := writer.Write(row); err != nil {
					return fmt.Errorf("could not write record to file: %v", err)
				}
			}
		case "http":
			// Write the header
			header := []string{
				"Type",
				"Seq",
				"Status",
				"Method",
				"URL",
				"Response_Code",
				"Response_Phase",
				"Response_Time(ms)",
				"SendDate",
				"SendTime",
				"SessionSent",
				"SessionSuccess",
				"FailureRate",
				"MinRtt",
				"AvgRtt",
				"MaxRtt",
				"AdditionalInfo",
			}

			err := writer.Write(header)

			if err != nil {
				return fmt.Errorf("could not write header to file: %v", err)
			}

			// Write each struct to the file
			for _, recordItem := range *dbTestEntries {

				// interface assertion
				pkt := recordItem.(*ntdb.RecordHTTPEntry)

				// url
				url := ntPinger.ConstructURL(iv.Http_scheme, iv.DestHost, iv.Http_path, iv.DestPort)

				// construct the row
				row := []string{
					iv.Type,                       // Ping Type
					strconv.Itoa(pkt.Seq),         // Seq
					fmt.Sprintf("%v", pkt.Status), // Status
					iv.Http_method,                // HTTP Method
					url,                           // DestHost
					pkt.ResponseCode,              // Response_Code
					pkt.ResponsePhase,             // Response Phase

					fmt.Sprintf("%v", float64(pkt.ResponseTime.Nanoseconds())/1e6), // Response_Time
					pkt.GetSendTime().Format("2006-01-02"),                         // SendDate
					pkt.GetSendTime().Format("15:04:05 MST"),                       // SendTime

					strconv.Itoa(pkt.Seq + 1),         // PacketsSent
					strconv.Itoa(pkt.SuccessResponse), // PacketsRecv
					pkt.FailRate,                      // PacketLoss
					pkt.MinRTT,                        // MinRtt
					pkt.AvgRTT,                        // AvgRtt
					pkt.MaxRTT,                        // MaxRtt
					pkt.AddInfo,                       // AdditionalInfo
				}

				if err := writer.Write(row); err != nil {
					return fmt.Errorf("could not write record to file: %v", err)
				}
			}
		case "dns":
			// Write the header
			header := []string{
				"Type",
				"Seq",
				"Status",
				"DNS_Resolver",
				"DNS_Query",
				"DNS_Response",
				"Record",
				"DNS_Protocol",
				"Response_Time(ms)",
				"SendDate",
				"SendTime",
				"PacketsSent",
				"SuccessResponse",
				"FailureRate",
				"MinRtt",
				"AvgRtt",
				"MaxRtt",
				"AdditionalInfo",
			}

			err := writer.Write(header)

			if err != nil {
				return fmt.Errorf("could not write header to file: %v", err)
			}

			// Write each struct to the file
			for _, recordItem := range *dbTestEntries {
				// interface assertion
				pkt := recordItem.(*ntdb.RecordDNSEntry)

				// construct the row
				row := []string{
					iv.Type,                       // Ping Type
					strconv.Itoa(pkt.Seq),         // Seq
					fmt.Sprintf("%v", pkt.Status), // Status
					iv.DestHost,                   // DNS_Resolver
					iv.Dns_query,                  // DNS_Query
					pkt.DnsResponse,               // DNS_Response
					pkt.DnsRecord,                 // Record
					iv.Dns_Protocol,               // DNS_Protocol

					fmt.Sprintf("%v", float64(pkt.ResponseTime.Nanoseconds())/1e6), // Response_Time
					pkt.GetSendTime().Format("2006-01-02"),                         // SendDate
					pkt.GetSendTime().Format("15:04:05 MST"),                       // SendTime

					strconv.Itoa(pkt.Seq + 1),         // PacketsSent
					strconv.Itoa(pkt.SuccessResponse), // PacketsRecv
					pkt.FailRate,                      // failure rate
					pkt.MinRTT,                        // MinRtt
					pkt.AvgRTT,                        // AvgRtt
					pkt.MaxRTT,                        // MaxRtt
					pkt.AddInfo,                       // AdditionalInfo
				}

				if err := writer.Write(row); err != nil {
					return fmt.Errorf("could not write record to file: %v", err)
				}
			}
		}
	}
	return nil
}

// GetDefaultExportFolder returns a folder path like ~/Documents/<appName>
// It creates the folder if it doesn't exist.
func GetDefaultExportFolder(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find user home directory: %w", err)
	}

	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = filepath.Join(home, "Documents")
	case "darwin": // macOS
		baseDir = filepath.Join(home, "Documents")
	case "linux":
		baseDir = filepath.Join(home, "Documents")
	default:
		// fallback if OS is unrecognized
		baseDir = filepath.Join(home, appName)
	}

	exportFolder := filepath.Join(baseDir, appName)

	// Ensure the directory exists
	if err := os.MkdirAll(exportFolder, 0755); err != nil {
		return "", fmt.Errorf("failed to create export folder: %w", err)
	}

	return exportFolder, nil
}
