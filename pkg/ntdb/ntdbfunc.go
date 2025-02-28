package ntdb

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"
)

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

// func: create Summary table if it does not exist
func EnsureSummaryTableExist(db *sql.DB) error {
	// Ensure summary table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS summary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start_datetime TEXT NOT NULL,
			test_type TEXT NOT NULL,
			test_uuID TEXT NOT NULL,
			command TEXT NOT NULL,
			recording_status BOOLEAN,
		);`)
	return err
}

// AddSummaryEntry inserts a new summary entry and creates a separate test result table if recording is enabled
func AddSummaryEntry(db *sql.DB, testUUID, testType, command string, recording bool) error {

	// ensure Summary Table exists
	err := EnsureSummaryTableExist(db)
	if err != nil {
		return err
	}

	// startTime
	startTime := time.Now().Format("2006-01-02 15:04:05")

	// recording satus
	recording_Status := ""
	if recording {
		recording_Status = "ON"
	} else {
		recording_Status = "OFF"
	}

	// Insert into summary table
	_, err = db.Exec(`INSERT INTO summary (start_datetime, test_type, test_uuID, command, recording_status) VALUES (?, ?, ?, ?, ?)`,
		startTime, testType, testUUID, command, recording_Status)
	if err != nil {
		return fmt.Errorf("failed to insert summary: %v", err)
	}
	return nil
}

// createTestResultsTable creates a unique test results table for each summary entry
func createTestResultsTable(db *sql.DB, testType, testUUID string) error {

	// generate table name
	tableName := fmt.Sprintf("table_%s", testUUID)

	// careate table based on test type
	switch testType {
	case "dns":
	case "http":
	case "tcp":
	case "icmp":
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		seq INTEGER,
		status TEXT,
		dns_resolver TEXT,
		dns_query TEXT,
		dns_response TEXT,
		record TEXT,
		response_time TEXT,
		send_datetime TEXT,
		success_response TEXT,
		failure_rate TEXT,
		min_rtt TEXT,
		max_rtt TEXT,
		avg_rtt TEXT,
		additional_info TEXT
	);`, tableName)

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}
	log.Println("Created test results table:", tableName)
	return nil
}

// AddTestResults inserts test results into a dynamically created test results table
// func AddTestResults(db *sql.DB, testUUID string, p *ntPinger.Packet) error {

// 	// obtain test Type
// 	testType := (*pkt).GetType()

// 	switch testType {
// 	case "dns":
// 		pkt := (*p).(*ntPinger.PacketDNS)

// 		query := fmt.Sprintf(`INSERT INTO %s (seq, status, target_host, rtt) VALUES (?, ?, ?, ?)`, tableName)
// 		_, err := db.Exec(query, res.Seq, res.Status, res.TargetHost, res.RTT)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert test result into %s: %v", tableName, err)
// 		}

// 	case "http":
// 	case "tcp":
// 	case "icmp":
// 	}

// 	for _, res := range results {
// 		query := fmt.Sprintf(`INSERT INTO %s (seq, status, target_host, rtt) VALUES (?, ?, ?, ?)`, tableName)
// 		_, err := db.Exec(query, res.Seq, res.Status, res.TargetHost, res.RTT)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert test result into %s: %v", tableName, err)
// 		}
// 	}
// 	log.Println("Successfully added test results to table:", tableName)
// 	return nil
// }
