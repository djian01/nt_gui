package testengine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var legacyDNSHeader = []string{"Type", "Seq", "Status", "DNS_Resolver", "DNS_Query", "DNS_Response", "Record", "DNS_Protocol", "Response_Time(ms)", "SendDate", "SendTime", "PacketsSent", "SuccessResponse", "FailureRate", "MinRtt", "AvgRtt", "MaxRtt", "AdditionalInfo"}
var dnsCSVHeader = append(append([]string{}, legacyDNSHeader...), "Test ID", "Interval (ms)", "Timeout (ms)", "Time (UTC)")

func dnsCSVRow(test, summary Session, p Sample) []string {
	duration := func(ms float64) string { return (time.Duration(ms * float64(time.Millisecond))).String() }
	return []string{
		"dns", strconv.Itoa(p.Sequence - 1), strconv.FormatBool(p.Success), test.Config.Resolver,
		test.Config.Query, p.DNSResponse, p.DNSRecord, test.Config.Protocol,
		strconv.FormatFloat(p.RTT, 'f', -1, 64), p.Time.UTC().Format("2006-01-02"), p.Time.UTC().Format("15:04:05 MST"),
		strconv.Itoa(summary.Sent), strconv.Itoa(summary.Succeeded),
		fmt.Sprintf("%.2f%%", float64(summary.Sent-summary.Succeeded)/float64(summary.Sent)*100),
		duration(summary.MinRTT), duration(summary.AvgRTT), duration(summary.MaxRTT), p.Error,
		test.ID, strconv.Itoa(test.Config.IntervalMS), strconv.Itoa(test.Config.TimeoutMS), p.Time.UTC().Format(time.RFC3339Nano),
	}
}

func parseDNSRow(record []string) (importedRow, error) {
	if strings.ToLower(strings.TrimSpace(record[0])) != "dns" {
		return importedRow{}, errors.New("CSV row is not a DNS result")
	}
	sequence, err := nonNegativeInt(record[1], "sequence")
	if err != nil {
		return importedRow{}, err
	}
	success, err := strconv.ParseBool(strings.TrimSpace(record[2]))
	if err != nil {
		return importedRow{}, errors.New("status must be true or false")
	}
	rtt, err := nonNegativeFloat(record[8], "response time")
	if err != nil {
		return importedRow{}, err
	}
	stamp, err := legacyTime(strings.TrimSpace(record[9]) + " " + strings.TrimSpace(record[10]))
	if err != nil {
		return importedRow{}, errors.New("invalid DNS send date or time")
	}
	sent, err := positiveInt(record[11], "packets sent")
	if err != nil {
		return importedRow{}, err
	}
	succeeded, err := nonNegativeInt(record[12], "success response")
	if err != nil || succeeded > sent {
		return importedRow{}, errors.New("success response must be between 0 and packets sent")
	}
	if _, err := parsePercent(record[13]); err != nil {
		return importedRow{}, err
	}
	minimum, err := durationMilliseconds(record[14], "minimum RTT")
	if err != nil {
		return importedRow{}, err
	}
	average, err := durationMilliseconds(record[15], "average RTT")
	if err != nil {
		return importedRow{}, err
	}
	maximum, err := durationMilliseconds(record[16], "maximum RTT")
	if err != nil {
		return importedRow{}, err
	}
	if succeeded > 0 && (minimum > average || average > maximum) {
		return importedRow{}, errors.New("DNS RTT summary is inconsistent")
	}
	c := Config{Type: "dns", Resolver: strings.TrimSpace(record[3]), Query: record[4],
		Protocol: strings.TrimSpace(record[7]), IntervalMS: 1000, TimeoutMS: 4000, Recording: true}
	sourceID := ""
	if len(record) == len(dnsCSVHeader) {
		sourceID = record[18]
		if c.IntervalMS, err = positiveInt(record[19], "interval"); err != nil {
			return importedRow{}, err
		}
		if c.TimeoutMS, err = positiveInt(record[20], "timeout"); err != nil {
			return importedRow{}, err
		}
		if stamp, err = time.Parse(time.RFC3339Nano, record[21]); err != nil {
			return importedRow{}, errors.New("invalid UTC probe time")
		}
	}
	c, err = validateDNS(c)
	if err != nil {
		return importedRow{}, err
	}
	return importedRow{sourceID: sourceID, config: c, sample: Sample{Sequence: sequence, Time: stamp, RTT: rtt,
		Success: success, DNSResponse: record[5], DNSRecord: record[6], Error: record[17]}}, nil
}
