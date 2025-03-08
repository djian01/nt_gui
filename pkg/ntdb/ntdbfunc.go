package ntdb

import (
	"database/sql"
	"fmt"
	"log"
)

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
