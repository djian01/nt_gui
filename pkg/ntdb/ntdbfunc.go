package ntdb

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
)

// createTestResultsTable creates a unique test results table for each summary entry
func CreateTestResultsTable(db *sql.DB, testType, testTableName string) error {

	// initial query
	query := ""

	// careate table based on test type
	switch testType {
	case "dns":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seq INTEGER,
			status TEXT,
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
		);`, testTableName)
	case "http":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seq INTEGER,
			status TEXT,
			response_code TEXT,
			response_phase TEXT,
			response_time TEXT,
			send_datetime TEXT,
			sessionsuccess TEXT,
			failure_rate TEXT,
			min_rtt TEXT,
			max_rtt TEXT,
			avg_rtt TEXT,
			additional_info TEXT
		);`, testTableName)
	case "tcp":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seq INTEGER,
			status TEXT,
			rtt TEXT,
			send_datetime TEXT,
			packetrecv TEXT,
			packetloss TEXT,
			min_rtt TEXT,
			max_rtt TEXT,
			avg_rtt TEXT,
			additional_info TEXT
		);`, testTableName)
	case "icmp":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seq INTEGER,
			status TEXT,
			RTT TEXT,
			send_datetime TEXT,
			packetrecv TEXT,
			packetloss TEXT,
			min_rtt TEXT,
			max_rtt TEXT,
			avg_rtt TEXT,
			additional_info TEXT
		);`, testTableName)
	}

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", testTableName, err)
	}

	return nil
}

// SortItems sorts the slice of Items by Index in ascending order.
func SortHistoryEntries(HistoryEntries *[]HistoryEntry) {
	sort.Slice(*HistoryEntries, func(i, j int) bool {
		indexI, errI := strconv.Atoi((*HistoryEntries)[i].Id)
		indexJ, errJ := strconv.Atoi((*HistoryEntries)[j].Id)

		// Handle conversion errors (place invalid indices at the end)
		if errI != nil || errJ != nil {
			return errI == nil // If errI is valid and errJ is invalid, keep it first
		}

		return indexI < indexJ // Sort by integer value of Index
	})
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
