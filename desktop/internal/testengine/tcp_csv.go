package testengine

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var legacyTCPHeader = []string{"Type", "Seq", "Status", "DestHost", "DestAddr", "DestPort", "PayLoadSize", "RTT(ms)", "SendDate", "SendTime", "PacketsSent", "PacketsRecv", "PacketLoss", "MinRtt", "AvgRtt", "MaxRtt", "AdditionalInfo"}
var tcpCSVHeader = append(append([]string{}, legacyTCPHeader...), "Test ID", "Interval (ms)", "Timeout (ms)", "Time (UTC)")

func tcpCSVRow(test, summary Session, p Sample) []string {
	duration := func(ms float64) string { return (time.Duration(ms * float64(time.Millisecond))).String() }
	return []string{
		"tcp", strconv.Itoa(p.Sequence - 1), strconv.FormatBool(p.Success), test.Config.Target,
		test.Config.ResolvedIP, strconv.Itoa(test.Config.Port), "0",
		strconv.FormatFloat(p.RTT, 'f', -1, 64), p.Time.UTC().Format("2006-01-02"), p.Time.UTC().Format("15:04:05 MST"),
		strconv.Itoa(summary.Sent), strconv.Itoa(summary.Succeeded),
		fmt.Sprintf("%.2f%%", float64(summary.Sent-summary.Succeeded)/float64(summary.Sent)*100),
		duration(summary.MinRTT), duration(summary.AvgRTT), duration(summary.MaxRTT), p.Error,
		test.ID, strconv.Itoa(test.Config.IntervalMS), strconv.Itoa(test.Config.TimeoutMS), p.Time.UTC().Format(time.RFC3339Nano),
	}
}

func parseTCPRow(record []string) (importedRow, error) {
	if strings.ToLower(strings.TrimSpace(record[0])) != "tcp" {
		return importedRow{}, errors.New("CSV row is not a TCP result")
	}
	sequence, err := nonNegativeInt(record[1], "sequence")
	if err != nil {
		return importedRow{}, err
	}
	success, err := strconv.ParseBool(strings.TrimSpace(record[2]))
	if err != nil {
		return importedRow{}, errors.New("status must be true or false")
	}
	rtt, err := nonNegativeFloat(record[7], "response time")
	if err != nil {
		return importedRow{}, err
	}
	stamp, err := legacyTime(strings.TrimSpace(record[8]) + " " + strings.TrimSpace(record[9]))
	if err != nil {
		return importedRow{}, errors.New("invalid TCP send date or time")
	}
	sent, err := positiveInt(record[10], "packets sent")
	if err != nil {
		return importedRow{}, err
	}
	succeeded, err := nonNegativeInt(record[11], "success response")
	if err != nil || succeeded > sent {
		return importedRow{}, errors.New("success response must be between 0 and packets sent")
	}
	if _, err := parsePercent(record[12]); err != nil {
		return importedRow{}, err
	}
	minimum, err := durationMilliseconds(record[13], "minimum RTT")
	if err != nil {
		return importedRow{}, err
	}
	average, err := durationMilliseconds(record[14], "average RTT")
	if err != nil {
		return importedRow{}, err
	}
	maximum, err := durationMilliseconds(record[15], "maximum RTT")
	if err != nil {
		return importedRow{}, err
	}
	if succeeded > 0 && (minimum > average || average > maximum) {
		return importedRow{}, errors.New("TCP RTT summary is inconsistent")
	}
	port, err := positiveInt(record[5], "TCP port")
	if err != nil {
		return importedRow{}, err
	}
	if payload, err := nonNegativeInt(record[6], "payload size"); err != nil || payload != 0 {
		return importedRow{}, errors.New("TCP connection tests require a zero payload size")
	}
	c := Config{Type: "tcp", Target: strings.TrimSpace(record[3]), Port: port,
		IntervalMS: 1000, TimeoutMS: 4000, Recording: true}
	// legacy sometimes wrote the hostname in both destination columns.
	if address := strings.TrimSpace(record[4]); net.ParseIP(address) != nil {
		c.ResolvedIP = address
	} else if address != c.Target && !(address == "" && len(record) == len(tcpCSVHeader)) {
		return importedRow{}, errors.New("Invalid TCP destination address")
	}
	sourceID := ""
	if len(record) == len(tcpCSVHeader) {
		sourceID = record[17]
		if c.IntervalMS, err = positiveInt(record[18], "interval"); err != nil {
			return importedRow{}, err
		}
		if c.TimeoutMS, err = positiveInt(record[19], "timeout"); err != nil {
			return importedRow{}, err
		}
		if stamp, err = time.Parse(time.RFC3339Nano, record[20]); err != nil {
			return importedRow{}, errors.New("invalid UTC probe time")
		}
	}
	c, err = validateTCP(c)
	if err != nil {
		return importedRow{}, err
	}
	return importedRow{sourceID: sourceID, config: c, sample: Sample{Sequence: sequence, Time: stamp, RTT: rtt,
		Success: success, Error: record[16]}}, nil
}
