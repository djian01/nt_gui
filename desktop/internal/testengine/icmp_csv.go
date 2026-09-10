package testengine

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var legacyICMPHeader = []string{"Type", "Seq", "Status", "DestHost", "DestAddr", "PayLoadSize", "RTT(ms)", "SendDate", "SendTime", "PacketsSent", "PacketsRecv", "PacketLoss", "MinRtt", "AvgRtt", "MaxRtt", "AdditionalInfo"}
var icmpCSVHeader = append(append([]string{}, legacyICMPHeader...), "Test ID", "Interval (ms)", "Timeout (ms)", "Time (UTC)", "DF")

func icmpCSVRow(test, summary Session, p Sample) []string {
	duration := func(ms float64) string { return (time.Duration(ms * float64(time.Millisecond))).String() }
	return []string{
		"icmp", strconv.Itoa(p.Sequence - 1), strconv.FormatBool(p.Success), test.Config.Target,
		test.Config.ResolvedIP, strconv.Itoa(test.Config.PayloadSize),
		strconv.FormatFloat(p.RTT, 'f', -1, 64), p.Time.UTC().Format("2006-01-02"), p.Time.UTC().Format("15:04:05 MST"),
		strconv.Itoa(summary.Sent), strconv.Itoa(summary.Succeeded),
		fmt.Sprintf("%.2f%%", float64(summary.Sent-summary.Succeeded)/float64(summary.Sent)*100),
		duration(summary.MinRTT), duration(summary.AvgRTT), duration(summary.MaxRTT), p.Error,
		test.ID, strconv.Itoa(test.Config.IntervalMS), strconv.Itoa(test.Config.TimeoutMS), p.Time.UTC().Format(time.RFC3339Nano), strconv.FormatBool(test.Config.DF),
	}
}

func parseICMPRow(record []string) (importedRow, error) {
	if strings.ToLower(strings.TrimSpace(record[0])) != "icmp" {
		return importedRow{}, errors.New("CSV row is not an ICMP result")
	}
	sequence, err := nonNegativeInt(record[1], "sequence")
	if err != nil {
		return importedRow{}, err
	}
	success, err := strconv.ParseBool(strings.TrimSpace(record[2]))
	if err != nil {
		return importedRow{}, errors.New("status must be true or false")
	}
	rtt, err := nonNegativeFloat(record[6], "response time")
	if err != nil {
		return importedRow{}, err
	}
	stamp, err := legacyTime(strings.TrimSpace(record[7]) + " " + strings.TrimSpace(record[8]))
	if err != nil {
		return importedRow{}, errors.New("invalid ICMP send date or time")
	}
	sent, err := positiveInt(record[9], "packets sent")
	if err != nil {
		return importedRow{}, err
	}
	succeeded, err := nonNegativeInt(record[10], "success response")
	if err != nil || succeeded > sent {
		return importedRow{}, errors.New("success response must be between 0 and packets sent")
	}
	if _, err := parsePercent(record[11]); err != nil {
		return importedRow{}, err
	}
	minimum, err := durationMilliseconds(record[12], "minimum RTT")
	if err != nil {
		return importedRow{}, err
	}
	average, err := durationMilliseconds(record[13], "average RTT")
	if err != nil {
		return importedRow{}, err
	}
	maximum, err := durationMilliseconds(record[14], "maximum RTT")
	if err != nil {
		return importedRow{}, err
	}
	if succeeded > 0 && (minimum > average || average > maximum) {
		return importedRow{}, errors.New("ICMP RTT summary is inconsistent")
	}
	payload, err := positiveInt(record[5], "payload size")
	if err != nil {
		return importedRow{}, err
	}
	c := Config{Type: "icmp", Target: strings.TrimSpace(record[3]), PayloadSize: payload,
		IntervalMS: 1000, TimeoutMS: 4000, Recording: true}
	// Fyne sometimes wrote the hostname in both destination columns.
	if address := strings.TrimSpace(record[4]); net.ParseIP(address) != nil {
		c.ResolvedIP = address
	} else if address != c.Target && !(address == "" && len(record) == len(icmpCSVHeader)) {
		return importedRow{}, errors.New("Invalid ICMP destination address")
	}
	sourceID := ""
	if len(record) == len(icmpCSVHeader) {
		sourceID = record[16]
		if c.DF, err = strconv.ParseBool(record[20]); err != nil {
			return importedRow{}, errors.New("DF must be true or false")
		}
		if c.IntervalMS, err = positiveInt(record[17], "interval"); err != nil {
			return importedRow{}, err
		}
		if c.TimeoutMS, err = positiveInt(record[18], "timeout"); err != nil {
			return importedRow{}, err
		}
		if stamp, err = time.Parse(time.RFC3339Nano, record[19]); err != nil {
			return importedRow{}, errors.New("invalid UTC probe time")
		}
	}
	c, err = validateICMP(c)
	if err != nil {
		return importedRow{}, err
	}
	return importedRow{sourceID: sourceID, config: c, sample: Sample{Sequence: sequence, Time: stamp, RTT: rtt,
		Success: success, Error: record[15]}}, nil
}
