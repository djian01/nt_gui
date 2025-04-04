package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

func SaveToCSV(filePath string, iv ntPinger.InputVars, dbEntries *[]ntdb.DbEntry) error {

	// set DateTime Layout
	dateTimeLayout := "2006-01-02 15:04:05 MST"

	// Open or create the file with append mode and write-only access
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// if accumulatedRecords is empty
	if len(*dbEntries) == 0 {
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
			for _, recordItem := range *dbEntries {
				// interface assertion
				pkt := recordItem.(*ntdb.RecordICMPEntry)

				// RTT
				RTT, err := strconv.ParseFloat(pkt.RTT, 64)
				if err != nil {
					return err
				}

				//DateTime
				t, err := time.Parse(pkt.SendDateTime, dateTimeLayout)
				if err != nil {
					return err
				}

				row := []string{
					iv.Type,                       // Ping Type
					strconv.Itoa(pkt.Seq),         // Seq
					fmt.Sprintf("%t", pkt.Status), // Status
					iv.DestHost,                   // DestHost
					iv.DestHost,                   // DestAddr
					strconv.Itoa(iv.PayLoadSize),  // PayLoadSize
					fmt.Sprintf("%v", RTT/1e6),    // Response_Time
					t.Format("2006-01-02"),        // SendDate
					t.Format("15:04:05 MST"),      // SendTime

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
			for _, recordItem := range accumulatedRecords {
				// interface assertion
				pkt := recordItem.(*ntPinger.PacketTCP)
				RTT := (float64((pkt.RTT).Nanoseconds())) / 1e6

				row := []string{
					pkt.Type,                            // Ping Type
					strconv.Itoa(pkt.Seq),               // Seq
					fmt.Sprintf("%t", pkt.Status),       // Status
					pkt.DestHost,                        // DestHost
					pkt.DestAddr,                        // DestAddr
					strconv.Itoa(pkt.DestPort),          // DestPort
					strconv.Itoa(pkt.PayLoadSize),       // PayLoadSize
					fmt.Sprintf("%v", RTT),              // Response_Time
					pkt.SendTime.Format("2006-01-02"),   // SendDate
					pkt.SendTime.Format("15:04:05 MST"), // SendTime

					strconv.Itoa(pkt.PacketsSent),                      // PacketsSent
					strconv.Itoa(pkt.PacketsRecv),                      // PacketsRecv
					fmt.Sprintf("%.2f%%", float64(pkt.PacketLoss*100)), // PacketLoss
					pkt.MinRtt.String(),                                // MinRtt
					pkt.AvgRtt.String(),                                // AvgRtt
					pkt.MaxRtt.String(),                                // MaxRtt
					pkt.AdditionalInfo,                                 // AdditionalInfo
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
			for _, recordItem := range accumulatedRecords {

				// interface assertion
				pkt := recordItem.(*ntPinger.PacketHTTP)
				RTT := (float64((pkt.RTT).Nanoseconds())) / 1e6

				// url
				url := ntPinger.ConstructURL(pkt.Http_scheme, pkt.DestHost, pkt.Http_path, pkt.DestPort)

				row := []string{
					pkt.Type,                             // Ping Type
					strconv.Itoa(pkt.Seq),                // Seq
					fmt.Sprintf("%t", pkt.Status),        // Status
					pkt.Http_method,                      // HTTP Method
					url,                                  // DestHost
					strconv.Itoa(pkt.Http_response_code), // Response_Code
					pkt.Http_response,                    // Response Phase
					fmt.Sprintf("%v", RTT),               // Response_Time
					pkt.SendTime.Format("2006-01-02"),    // SendDate
					pkt.SendTime.Format("15:04:05 MST"),  // SendTime

					strconv.Itoa(pkt.PacketsSent),                      // PacketsSent
					strconv.Itoa(pkt.PacketsRecv),                      // PacketsRecv
					fmt.Sprintf("%.2f%%", float64(pkt.PacketLoss*100)), // PacketLoss
					pkt.MinRtt.String(),                                // MinRtt
					pkt.AvgRtt.String(),                                // AvgRtt
					pkt.MaxRtt.String(),                                // MaxRtt
					pkt.AdditionalInfo,                                 // AdditionalInfo
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
			for _, recordItem := range accumulatedRecords {
				// interface assertion
				pkt := recordItem.(*ntPinger.PacketDNS)
				RTT := (float64((pkt.RTT).Nanoseconds())) / 1e6

				row := []string{
					pkt.Type,                            // Ping Type
					strconv.Itoa(pkt.Seq),               // Seq
					fmt.Sprintf("%t", pkt.Status),       // Status
					pkt.DestHost,                        // DNS_Resolver
					pkt.Dns_query,                       // DNS_Query
					pkt.Dns_response,                    // DNS_Response
					pkt.Dns_queryType,                   // Record
					pkt.Dns_protocol,                    // DNS_Protocol
					fmt.Sprintf("%v", RTT),              // Response_Time
					pkt.SendTime.Format("2006-01-02"),   // SendDate
					pkt.SendTime.Format("15:04:05 MST"), // SendTime
					strconv.Itoa(pkt.PacketsSent),       // PacketsSent
					strconv.Itoa(pkt.PacketsRecv),       // PacketsRecv
					fmt.Sprintf("%.2f%%", float64(pkt.PacketLoss*100)), // PacketLoss
					pkt.MinRtt.String(), // MinRtt
					pkt.AvgRtt.String(), // AvgRtt
					pkt.MaxRtt.String(), // MaxRtt
					pkt.AdditionalInfo,  // AdditionalInfo
				}

				if err := writer.Write(row); err != nil {
					return fmt.Errorf("could not write record to file: %v", err)
				}
			}
		}
	}
	return nil
}
