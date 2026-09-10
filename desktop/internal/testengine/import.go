package testengine

import (
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	maxImportBytes = 256 << 20
	maxImportRows  = 1_000_000
)

var (
	wailsCSVHeader     = []string{"Test ID", "URL", "Method", "Interval (ms)", "Timeout (ms)", "Sequence", "Time (UTC)", "RTT (ms)", "HTTP status", "Success", "Error"}
	wailsCSVFullHeader = append(append([]string{}, wailsCSVHeader...), "Response phase", "Expected statuses", "Proxy URL", "Proxy username", "Proxy password required", "Follow redirects")
	legacyHTTPHeader   = []string{"Type", "Seq", "Status", "Method", "URL", "Response_Code", "Response_Phase", "Response_Time(ms)", "SendDate", "SendTime", "SessionSent", "SessionSuccess", "FailureRate", "MinRtt", "AvgRtt", "MaxRtt", "AdditionalInfo"}
)

type importFormat int

const (
	importWails importFormat = iota
	importLegacyHTTP
	importDNS
	importTCP
	importICMP
)

type importedRow struct {
	sourceID         string
	config           Config
	passwordRequired bool
	sample           Sample
}

// ImportCSV imports an HTTP, DNS, TCP, or ICMP test from a desktop or legacy
// CSV export. The imported test gets a fresh identity and is stopped.
// Parsing and database writes are streamed, and any invalid row rolls back the
// entire import.
func (s *Store) ImportCSV(input io.Reader) (Session, error) {
	limited := &io.LimitedReader{R: input, N: maxImportBytes + 1}
	reader := csv.NewReader(limited)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Session{}, errors.New("CSV file is empty")
		}
		return Session{}, fmt.Errorf("Could not read CSV header: %w", err)
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	format, err := detectImportFormat(header)
	if err != nil {
		return Session{}, err
	}
	reader.FieldsPerRecord = len(header)

	tx, err := s.db.Begin()
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()

	var session Session
	var sourceID string
	var previousTime time.Time
	var originalSequence int
	accepted := make(map[int]bool)
	rows := 0
	initialized := false
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return Session{}, fmt.Errorf("Invalid CSV row %d: %w", rows+2, readErr)
		}
		rows++
		if rows > maxImportRows {
			return Session{}, fmt.Errorf("CSV contains more than %d probe rows", maxImportRows)
		}
		row, parseErr := parseImportedRow(format, record)
		if parseErr != nil {
			return Session{}, fmt.Errorf("Invalid CSV row %d: %w", rows+1, parseErr)
		}
		if !initialized {
			session = Session{
				ID:               rand.Text(),
				Config:           row.config,
				StartedAt:        row.sample.Time,
				Revision:         1,
				EndReason:        "imported",
				PasswordRequired: row.passwordRequired,
			}
			sourceID = row.sourceID
			originalSequence = row.sample.Sequence
			if format == importICMP {
				session.ImportNote = "ICMP CSV: statistics cover recorded rows only; sequences are renumbered for analysis."
				if len(header) == len(legacyICMPHeader) {
					session.ImportNote += " Legacy replay uses a 1-second interval, 4-second timeout, and DF OFF because these settings were not stored."
				}
			} else if format == importTCP {
				session.ImportNote = "TCP CSV: statistics cover recorded rows only; sequences are renumbered for analysis."
				if len(header) == len(legacyTCPHeader) {
					session.ImportNote += " Legacy replay uses a 1-second interval and 4-second timeout; unresolved legacy targets are resolved when replayed."
				}
			} else if format == importDNS {
				originalSequence = row.sample.Sequence
				session.ImportNote = "DNS CSV: statistics cover recorded rows only; sequences are renumbered for analysis."
				if len(header) == len(legacyDNSHeader) {
					session.ImportNote += " Legacy replay uses a 1-second interval and 4-second timeout because those settings were not stored."
				}
			} else if format == importLegacyHTTP {
				originalSequence = row.sample.Sequence
				session.ImportNote = fmt.Sprintf("Legacy HTTP CSV: sequences renumbered from original #%d. Statistics cover imported rows only. Replay uses a 1-second interval, 4-second timeout, follows redirects, and has no proxy; expected statuses are inferred from saved successes. Original replay settings were not stored in this format.", originalSequence)
			} else if len(header) == len(wailsCSVHeader) {
				session.ImportNote = "Older desktop CSV: proxy and expected status settings were not stored. Replay uses no proxy and expected statuses inferred from saved successes; redirects are not followed."
			}
			if format == importWails && originalSequence != 1 {
				session.ImportNote += " Partial recording: sequences renumbered; statistics cover recorded rows only."
			}
			if err := insertImportedSession(tx, session); err != nil {
				return Session{}, err
			}
			initialized = true
		} else {
			if err := consistentImport(format, sourceID, session.Config, row); err != nil {
				return Session{}, fmt.Errorf("Invalid CSV row %d: %w", rows+1, err)
			}
			if session.PasswordRequired != row.passwordRequired {
				return Session{}, errors.New("Proxy password requirement changes between rows")
			}
		}
		if row.sample.Sequence-originalSequence != rows-1 {
			return Session{}, fmt.Errorf("Invalid CSV row %d: probe sequences must be consecutive", rows+1)
		}
		row.sample.Sequence = rows
		if !previousTime.IsZero() && row.sample.Time.Before(previousTime) {
			return Session{}, fmt.Errorf("Invalid CSV row %d: probe times must be chronological", rows+1)
		}
		previousTime = row.sample.Time
		if row.sample.Success && row.sample.StatusCode > 0 {
			accepted[row.sample.StatusCode] = true
		}
		updateImportedSummary(&session, row.sample)
		if err := saveSample(tx, session.ID, &row.sample); err != nil {
			return Session{}, err
		}
	}
	if limited.N == 0 {
		return Session{}, fmt.Errorf("CSV exceeds the %d MiB import limit", maxImportBytes>>20)
	}
	if rows == 0 {
		return Session{}, errors.New("CSV contains no probe rows")
	}
	if session.Config.Type != "dns" && session.Config.Type != "tcp" && session.Config.Type != "icmp" && len(session.Config.AcceptedStatuses) == 0 {
		session.Config.AcceptedStatuses = inferredStatuses(accepted)
	}
	ended := previousTime
	session.EndedAt = &ended
	session.Revision = session.Sent + 1
	data, err := json.Marshal(session)
	if err != nil {
		return Session{}, err
	}
	if _, err = tx.Exec("UPDATE tests SET running=0,data=? WHERE id=?", string(data), session.ID); err != nil {
		return Session{}, fmt.Errorf("Could not finalize imported test: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("Could not save imported test: %w", err)
	}
	return session, nil
}

func detectImportFormat(header []string) (importFormat, error) {
	if equalHeader(header, legacyICMPHeader) || equalHeader(header, icmpCSVHeader) {
		return importICMP, nil
	}
	if equalHeader(header, legacyTCPHeader) || equalHeader(header, tcpCSVHeader) {
		return importTCP, nil
	}
	if equalHeader(header, legacyDNSHeader) || equalHeader(header, dnsCSVHeader) {
		return importDNS, nil
	}
	if equalHeader(header, wailsCSVHeader) || equalHeader(header, wailsCSVFullHeader) {
		return importWails, nil
	}
	if equalHeader(header, legacyHTTPHeader) {
		return importLegacyHTTP, nil
	}
	return 0, errors.New("Unsupported CSV: choose a NET-Test desktop export or a legacy HTTP/DNS/TCP/ICMP export")
}

func equalHeader(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if strings.TrimSpace(got[i]) != want[i] {
			return false
		}
	}
	return true
}

func parseImportedRow(format importFormat, record []string) (importedRow, error) {
	if format == importICMP {
		return parseICMPRow(record)
	}
	if format == importTCP {
		return parseTCPRow(record)
	}
	if format == importDNS {
		return parseDNSRow(record)
	}
	if format == importWails {
		return parseWailsRow(record)
	}
	return parseLegacyHTTPRow(record)
}

func parseWailsRow(record []string) (importedRow, error) {
	interval, err := positiveInt(record[3], "interval")
	if err != nil {
		return importedRow{}, err
	}
	timeout, err := positiveInt(record[4], "timeout")
	if err != nil {
		return importedRow{}, err
	}
	sequence, err := positiveInt(record[5], "sequence")
	if err != nil {
		return importedRow{}, err
	}
	stamp, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(record[6]))
	if err != nil {
		return importedRow{}, errors.New("invalid UTC probe time")
	}
	rtt, err := nonNegativeFloat(record[7], "RTT")
	if err != nil {
		return importedRow{}, err
	}
	status, err := nonNegativeInt(record[8], "HTTP status")
	if err != nil || status > 999 {
		return importedRow{}, errors.New("HTTP status must be between 0 and 999")
	}
	success, err := strconv.ParseBool(strings.TrimSpace(record[9]))
	if err != nil {
		return importedRow{}, errors.New("success must be true or false")
	}
	config := Config{URL: strings.TrimSpace(record[1]), Method: strings.ToUpper(strings.TrimSpace(record[2])), IntervalMS: interval, TimeoutMS: timeout}
	passwordRequired := false
	responsePhase := ""
	if len(record) == len(wailsCSVFullHeader) {
		responsePhase = record[11]
		if strings.TrimSpace(record[12]) != "" {
			config.AcceptedStatuses = splitMetadataList(record[12])
			var statusErr error
			config.AcceptedStatuses, _, statusErr = validateStatuses(config.AcceptedStatuses)
			if statusErr != nil {
				return importedRow{}, statusErr
			}
		}
		config.Proxy.URL = strings.TrimSpace(record[13])
		config.Proxy.Username = strings.TrimSpace(record[14])
		config.Proxy.Enabled = config.Proxy.URL != ""
		passwordRequired, err = strconv.ParseBool(strings.TrimSpace(record[15]))
		if err != nil {
			return importedRow{}, errors.New("proxy password required must be true or false")
		}
		config.FollowRedirects, err = strconv.ParseBool(strings.TrimSpace(record[16]))
		if err != nil {
			return importedRow{}, errors.New("follow redirects must be true or false")
		}
	}
	if err := validateImportedConfig(config); err != nil {
		return importedRow{}, err
	}
	if passwordRequired && (!config.Proxy.Enabled || config.Proxy.Username == "") {
		return importedRow{}, errors.New("Password-required proxy needs its URL and username")
	}
	if success && status == 0 {
		return importedRow{}, errors.New("Successful HTTP probe requires a response code")
	}
	return importedRow{
		sourceID: record[0], config: config, passwordRequired: passwordRequired,
		sample: Sample{Sequence: sequence, Time: stamp, RTT: rtt, StatusCode: status, Success: success, Error: record[10], ResponsePhase: responsePhase},
	}, nil
}

func splitMetadataList(input string) []string {
	var result []string
	for _, value := range strings.Split(input, ",") {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func parseLegacyHTTPRow(record []string) (importedRow, error) {
	if strings.ToLower(strings.TrimSpace(record[0])) != "http" {
		return importedRow{}, errors.New("legacy CSV row is not an HTTP result")
	}
	legacySeq, err := nonNegativeInt(record[1], "sequence")
	if err != nil {
		return importedRow{}, errors.New("legacy probe sequence must be non-negative")
	}
	success, err := strconv.ParseBool(strings.TrimSpace(record[2]))
	if err != nil {
		return importedRow{}, errors.New("status must be true or false")
	}
	status, err := nonNegativeInt(record[5], "response code")
	if err != nil || status > 999 {
		return importedRow{}, errors.New("response code must be between 0 and 999")
	}
	rtt, err := nonNegativeFloat(record[7], "response time")
	if err != nil {
		return importedRow{}, err
	}
	stamp, err := legacyTime(strings.TrimSpace(record[8]) + " " + strings.TrimSpace(record[9]))
	if err != nil {
		return importedRow{}, errors.New("invalid legacy send date or time")
	}
	sent, err := positiveInt(record[10], "session sent")
	if err != nil {
		return importedRow{}, err
	}
	succeeded, err := nonNegativeInt(record[11], "session success")
	if err != nil || succeeded > sent {
		return importedRow{}, errors.New("session success must be between 0 and session sent")
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
		return importedRow{}, errors.New("legacy RTT summary is inconsistent")
	}
	config := Config{URL: strings.TrimSpace(record[4]), Method: strings.ToUpper(strings.TrimSpace(record[3])), IntervalMS: 1000, TimeoutMS: 4000, FollowRedirects: true}
	if err := validateImportedConfig(config); err != nil {
		return importedRow{}, err
	}
	errorText := record[16]
	if errorText == "" && !success {
		errorText = record[6]
	}
	return importedRow{
		config: config,
		sample: Sample{Sequence: legacySeq, Time: stamp, RTT: rtt, StatusCode: status, Success: success, Error: errorText, ResponsePhase: record[6]},
	}, nil
}

func validateImportedConfig(config Config) error {
	if len(config.AcceptedStatuses) == 0 {
		config.AcceptedStatuses = []string{"2xx", "3xx"}
	}
	_, err := validate(config)
	return err
}

func consistentImport(format importFormat, sourceID string, config Config, row importedRow) error {
	if (format == importWails || format == importDNS || format == importTCP || format == importICMP) && row.sourceID != sourceID {
		return errors.New("CSV contains more than one test ID")
	}
	if !reflect.DeepEqual(row.config, config) {
		return errors.New("test metadata changes between rows")
	}
	return nil
}

func updateImportedSummary(session *Session, sample Sample) {
	session.Sent++
	copy := sample
	session.Last = &copy
	if !sample.Success {
		return
	}
	session.Succeeded++
	if session.Succeeded == 1 || sample.RTT < session.MinRTT {
		session.MinRTT = sample.RTT
	}
	if sample.RTT > session.MaxRTT {
		session.MaxRTT = sample.RTT
	}
	session.AvgRTT += (sample.RTT - session.AvgRTT) / float64(session.Succeeded)
}

func inferredStatuses(codes map[int]bool) []string {
	result := make([]string, 0, len(codes))
	for code := 200; code <= 599; code++ {
		if codes[code] {
			result = append(result, strconv.Itoa(code))
		}
	}
	if len(result) == 0 || len(result) > 20 {
		return []string{"2xx", "3xx"}
	}
	return result
}

func insertImportedSession(tx *sql.Tx, session Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO tests(id,url,running,data) VALUES(?,?,0,?)", session.ID, session.Config.URL, string(data))
	if err != nil {
		return fmt.Errorf("Could not create imported test: %w", err)
	}
	return nil
}

func positiveInt(value, name string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return n, nil
}

func nonNegativeInt(value, name string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return n, nil
}

func nonNegativeFloat(value, name string) (float64, error) {
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("%s must be a finite non-negative number", name)
	}
	return n, nil
}

func durationMilliseconds(value, name string) (float64, error) {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || d < 0 {
		return 0, fmt.Errorf("%s is invalid", name)
	}
	return float64(d) / float64(time.Millisecond), nil
}

func parsePercent(value string) (float64, error) {
	n, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(value), "%"), 64)
	if err != nil || n < 0 || n > 100 || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, errors.New("failure rate must be between 0% and 100%")
	}
	return n, nil
}

func legacyTime(value string) (time.Time, error) {
	const layout = "2006-01-02 15:04:05 MST"
	stamp, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	zone, offset := stamp.Zone()
	if offset != 0 || zone == "UTC" || zone == "GMT" {
		return stamp, nil
	}
	// Go otherwise invents a zero-offset zone for unknown abbreviations.
	offsets := map[string]int{"AEST": 600, "AEDT": 660, "ACST": 570, "ACDT": 630, "AWST": 480, "NZST": 720, "NZDT": 780, "PST": -480, "PDT": -420, "MST": -420, "MDT": -360, "CST": -360, "CDT": -300, "EST": -300, "EDT": -240, "CET": 60, "CEST": 120, "BST": 60}
	minutes, ok := offsets[zone]
	if !ok {
		return time.Time{}, fmt.Errorf("Unknown legacy time zone %q", zone)
	}
	return time.ParseInLocation(layout, value, time.FixedZone(zone, minutes*60))
}
