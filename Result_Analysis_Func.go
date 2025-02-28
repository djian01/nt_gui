package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ntHttp "github.com/djian01/nt/pkg/cmd/http"
	"github.com/djian01/nt/pkg/ntPinger"

	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

// func: Append Packet Slide
func appendPacket(inputResultPackets *[]ntPinger.Packet, RaType string, records *[][]string, chartData *[]ntchart.ChartPoint, Summary *Summary) {

	var chartPoint ntchart.ChartPoint
	recordLen := len(*records)

	switch RaType {
	case "dns":
		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketDNS
			p.Type = packet[0]
			p.Seq, _ = strconv.Atoi(packet[1])
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.DestAddr = packet[3]
			p.DestHost = packet[3]
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05 MST", packet[9]+" "+packet[10])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[8] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.Dns_query = packet[4]
			p.Dns_queryType = packet[6]
			p.Dns_protocol = packet[7]
			p.Dns_response = packet[5]
			p.PacketsSent, _ = strconv.Atoi(packet[11])
			p.PacketsRecv, _ = strconv.Atoi(packet[12])
			p.MinRtt, _ = parseCustomDuration(packet[14])
			p.MaxRtt, _ = parseCustomDuration(packet[16])
			p.AvgRtt, _ = parseCustomDuration(packet[15])
			index_AdditionalInfo := 17
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// update summary
			//// if it's the 1st packet
			if i == 1 {
				(*Summary).StartTime = p.SendTime
				(*Summary).DestHost = p.DestHost
				(*Summary).ntCmd = RaNtCmdGenerator(p.Type, &p)
			}
			//// if it's the last packet
			if i == recordLen-1 {
				(*Summary).EndTime = p.SendTime
				(*Summary).PacketSent = p.PacketsSent
				(*Summary).SuccessResponse = p.PacketsRecv
				(*Summary).FailRate = packet[13]
				(*Summary).MinRTT = p.MinRtt
				(*Summary).MaxRTT = p.MaxRtt
				(*Summary).AvgRtt = p.AvgRtt
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)

		}
	case "http":

		RAHttpVar, _ := ntHttp.ParseURL((*records)[1][4])

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketHTTP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestHost = RAHttpVar.Hostname
			p.DestPort = RAHttpVar.Port
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05 MST", packet[8]+" "+packet[9])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[7] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.Http_path = RAHttpVar.Path
			p.Http_scheme = RAHttpVar.Scheme
			p.Http_response_code, _ = strconv.Atoi(packet[5])
			p.Http_response = packet[6]
			p.Http_method = packet[3]
			p.PacketsSent, _ = strconv.Atoi(packet[10])
			p.PacketsRecv, _ = strconv.Atoi(packet[11])
			p.MinRtt, _ = parseCustomDuration(packet[13])
			p.MaxRtt, _ = parseCustomDuration(packet[15])
			p.AvgRtt, _ = parseCustomDuration(packet[14])
			index_AdditionalInfo := 16
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// update summary
			//// if it's the 1st packet
			if i == 1 {
				(*Summary).StartTime = p.SendTime
				(*Summary).DestHost = p.DestHost
				(*Summary).ntCmd = RaNtCmdGenerator(p.Type, &p)
			}
			//// if it's the last packet
			if i == recordLen-1 {
				(*Summary).EndTime = p.SendTime
				(*Summary).PacketSent = p.PacketsSent
				(*Summary).SuccessResponse = p.PacketsRecv
				(*Summary).FailRate = packet[12]
				(*Summary).MinRTT = p.MinRtt
				(*Summary).MaxRTT = p.MaxRtt
				(*Summary).AvgRtt = p.AvgRtt
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}

	case "tcp":

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketTCP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestAddr = packet[4]
			p.DestHost = packet[3]
			p.DestPort, _ = strconv.Atoi(packet[5])
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05 MST", packet[8]+" "+packet[9])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[7] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.PayLoadSize, _ = strconv.Atoi(packet[6])
			p.PacketsSent, _ = strconv.Atoi(packet[10])
			p.PacketsRecv, _ = strconv.Atoi(packet[11])
			p.MinRtt, _ = parseCustomDuration(packet[13])
			p.MaxRtt, _ = parseCustomDuration(packet[15])
			p.AvgRtt, _ = parseCustomDuration(packet[14])
			index_AdditionalInfo := 16
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// update summary
			//// if it's the 1st packet
			if i == 1 {
				(*Summary).StartTime = p.SendTime
				(*Summary).DestHost = p.DestHost
				(*Summary).ntCmd = RaNtCmdGenerator(p.Type, &p)
			}
			//// if it's the last packet
			if i == recordLen-1 {
				(*Summary).EndTime = p.SendTime
				(*Summary).PacketSent = p.PacketsSent
				(*Summary).SuccessResponse = p.PacketsRecv
				(*Summary).FailRate = packet[12]
				(*Summary).MinRTT = p.MinRtt
				(*Summary).MaxRTT = p.MaxRtt
				(*Summary).AvgRtt = p.AvgRtt
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}

	case "icmp":

		for i, packet := range *records {
			// skip the header row
			if i == 0 {
				continue
			}

			// fill in packet info
			var p ntPinger.PacketICMP
			p.Type = packet[0]
			p.Status, _ = strconv.ParseBool(packet[2])
			chartPoint.Status = p.Status
			p.Seq, _ = strconv.Atoi(packet[1])
			p.DestAddr = packet[4]
			p.DestHost = packet[3]
			p.PayLoadSize, _ = strconv.Atoi(packet[5])
			p.SendTime, _ = time.Parse("2006-01-02 15:04:05 MST", packet[7]+" "+packet[8])
			chartPoint.XValues = p.SendTime
			p.RTT, _ = parseCustomDuration(packet[6] + "ms")
			chartPoint.YValues = float64(p.RTT) / float64(time.Millisecond)
			p.PacketsSent, _ = strconv.Atoi(packet[9])
			p.PacketsRecv, _ = strconv.Atoi(packet[10])
			p.MinRtt, _ = parseCustomDuration(packet[12])
			p.MaxRtt, _ = parseCustomDuration(packet[14])
			p.AvgRtt, _ = parseCustomDuration(packet[13])
			index_AdditionalInfo := 15
			if len(packet) == index_AdditionalInfo+1 {
				p.AdditionalInfo = packet[index_AdditionalInfo]
			} else {
				p.AdditionalInfo = ""
			}

			// update summary
			//// if it's the 1st packet
			if i == 1 {
				(*Summary).StartTime = p.SendTime
				(*Summary).DestHost = p.DestHost
				(*Summary).ntCmd = RaNtCmdGenerator(p.Type, &p)
			}
			//// if it's the last packet
			if i == recordLen-1 {
				(*Summary).EndTime = p.SendTime
				(*Summary).PacketSent = p.PacketsSent
				(*Summary).SuccessResponse = p.PacketsRecv
				(*Summary).FailRate = packet[11]
				(*Summary).MinRTT = p.MinRtt
				(*Summary).MaxRTT = p.MaxRtt
				(*Summary).AvgRtt = p.AvgRtt
			}

			// append
			*inputResultPackets = append(*inputResultPackets, &p)
			*chartData = append(*chartData, chartPoint)
		}
	}
}

// Parse a duration string with "ms" (milliseconds) or "s" (seconds) to time.Duration
func parseCustomDuration(input string) (time.Duration, error) {

	var multiplier float64

	// Check the suffix and set the multiplier accordingly
	switch {
	case strings.HasSuffix(input, "ms"):
		multiplier = float64(time.Millisecond)
		input = strings.TrimSuffix(input, "ms") // Remove the "ms" suffix
	case strings.HasSuffix(input, "s"):
		multiplier = float64(time.Second)
		input = strings.TrimSuffix(input, "s") // Remove the "s" suffix
	default:
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", input)
	}

	// Use the multiplier to compute the duration
	duration := time.Duration(value * multiplier)
	return duration, nil
}

// NT CMD Generator
func RaNtCmdGenerator(RaType string, pk ntPinger.Packet) string {
	ntCmd := ""

	switch RaType {
	case "dns":
		myPk := *(pk.(*ntPinger.PacketDNS))
		if myPk.Dns_protocol == "tcp" {
			ntCmd = fmt.Sprintf("nt -r dns -o tcp %v %v", myPk.DestHost, myPk.Dns_query)
		} else {
			ntCmd = fmt.Sprintf("nt -r dns %v %v", myPk.DestHost, myPk.Dns_query)
		}

	case "http":
		myPk := *(pk.(*ntPinger.PacketHTTP))
		httpUrl := ntPinger.ConstructURL(myPk.Http_scheme, myPk.DestHost, myPk.Http_path, myPk.DestPort)

		if myPk.Http_method != "GET" {
			ntCmd = fmt.Sprintf("nt -r http -m %v %v", myPk.Http_method, httpUrl)
		} else {
			ntCmd = fmt.Sprintf("nt -r http %v ", httpUrl)
		}

	case "tcp":
		myPk := *(pk.(*ntPinger.PacketTCP))
		ntCmd = fmt.Sprintf("nt -r tcp %v %v", myPk.DestAddr, myPk.DestPort)

	case "icmp":
		myPk := *(pk.(*ntPinger.PacketICMP))
		if myPk.PayLoadSize > 32 {
			ntCmd = fmt.Sprintf("nt -r icmp -s %v %v", myPk.PayLoadSize, myPk.DestAddr)
		} else {
			ntCmd = fmt.Sprintf("nt -r icmp %v", myPk.DestAddr)
		}
	}

	return ntCmd
}
