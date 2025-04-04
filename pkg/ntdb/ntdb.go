package ntdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/djian01/nt/pkg/ntPinger"
	_ "modernc.org/sqlite" // Import SQLite driver
)

func DBOpen(dbFile string) (*sql.DB, error) {

	// check "dbFile" in the same folder as the executable
	// if db file not exist, os.Stat(dbFile) return error. And os.IsNotExist(err) returns true if err exist
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		// fmt.Println("Database file not found, creating new database...")
		err := createDatabase(dbFile)
		if err != nil {
			return nil, errors.New("failed to create database")
		}
		// fmt.Println("Database created successfully!")
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, errors.New("failed to open database")
	}

	return db, nil
}

// createDatabase creates a new SQLite database file, enable Auto VACUUM, and create the default History Table
func createDatabase(dbFile string) error {
	// Create an empty database file
	file, err := os.Create(dbFile)
	if err != nil {
		return err
	}
	file.Close() // Close immediately since SQLite will manage it

	// Open SQLite connection
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// Enable auto_vacuum mode
	_, err = db.Exec("PRAGMA auto_vacuum = FULL;")
	if err != nil {
		return fmt.Errorf("failed to set auto_vacuum: %v", err)
	}

	// Important: run VACUUM to activate auto_vacuum
	_, err = db.Exec("VACUUM;")
	if err != nil {
		return fmt.Errorf("failed to vacuum after setting auto_vacuum: %v", err)
	}

	// Create a default table (example: History)
	return createHistoryTable(db)
}

// createDefaultTable creates that a "users" table exists in the database
func createHistoryTable(db *sql.DB) error {

	// default table name is "history"
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tablename TEXT NOT NULL,
		testtype TEXT NOT NULL,
		starttime TEXT NOT NULL,
		command TEXT NOT NULL,
		uuid TEXT NOT NULL,
		recorded INTEGER NOT NULL DEFAULT 0
	);`
	_, err := db.Exec(query)
	return err
}

// func: convert *pkt to DbEntry
func ConvertPkt2DbEntry(pkt ntPinger.Packet, tableName string) (dbEntry DbEntry) {

	// get test type of pkt
	testType := strings.ToLower(pkt.GetType())

	// construct dbEntry based on test type
	switch testType {
	case "dns":
		testPkt := (pkt).(*ntPinger.PacketDNS)
		dnsEntry := RecordDNSEntry{}
		dnsEntry.TableName = tableName
		dnsEntry.TestType = testType
		dnsEntry.Seq = (*testPkt).Seq
		dnsEntry.Status = strconv.FormatBool((*testPkt).Status)
		dnsEntry.DnsResponse = (*testPkt).Dns_response
		dnsEntry.DnsRecord = (*testPkt).Dns_queryType
		dnsEntry.ResponseTime = fmt.Sprintf("%v", float64((*testPkt).RTT.Nanoseconds()))
		dnsEntry.SendDateTime = (*testPkt).SendTime.Format("2006-01-02 15:04:05 MST")
		dnsEntry.SuccessResponse = (*testPkt).PacketsRecv
		dnsEntry.FailRate = fmt.Sprintf("%.2f%%", float64((*testPkt).PacketLoss*100))
		dnsEntry.MinRTT = (*testPkt).MinRtt.String()
		dnsEntry.MaxRTT = (*testPkt).MaxRtt.String()
		dnsEntry.AvgRTT = (*testPkt).AvgRtt.String()
		dnsEntry.AddInfo = (*testPkt).AdditionalInfo
		dbEntry = &dnsEntry
	case "http":
	case "tcp":
	case "icmp":

	}

	return
}

// InsertEntry inserts a log entry into the "history" table
func InsertEntry(ntdb *sql.DB, entryChan <-chan DbEntry, errChan chan error) {

	// initial err
	var err error = nil

	// read from channel
	for entry := range entryChan {
		tableName := entry.GetTableName()
		switch tableName {
		// history table
		case "history":
			he := entry.(*HistoryEntry)
			// Construct SQL query with the dynamic table name
			query := fmt.Sprintf("INSERT INTO %s (tablename, testtype, starttime, command, uuid, recorded) VALUES (?, ?, ?, ?, ?, ?);", tableName)

			// setup temporary variable for recorded
			var recordedInt int // temporary variable to store the INT value of recorded
			if he.Recorded {
				recordedInt = 1
			} else {
				recordedInt = 0
			}

			// Execute the query safely with placeholders for values
			for {
				_, err = ntdb.Exec(query, he.TableName, he.TestType, he.StartTime, he.Command, he.UUID, recordedInt)
				if err != nil {
					// handle the "database is locked" error
					if strings.Contains(err.Error(), "database is locked") {
						time.Sleep(time.Millisecond * 100)
					} else {
						errChan <- err
						break
					}
				} else {
					break
				}
			}

		default:
			// recording table name example "dns_U4S2CP"
			tableNameSlice := strings.Split(tableName, "_")

			if len(tableNameSlice) == 2 {
				switch tableNameSlice[0] {
				case "dns":
					en := entry.(*RecordDNSEntry)
					// Construct SQL query with the dynamic table name
					query := fmt.Sprintf("INSERT INTO %s (seq, status, dns_response, record, response_time, send_datetime, success_response, failure_rate, min_rtt, max_rtt, avg_rtt, additional_info) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", tableName)

					// Execute the query safely with placeholders for values
					for {
						_, err = ntdb.Exec(query, en.Seq, en.Status, en.DnsResponse, en.DnsRecord, en.ResponseTime, en.SendDateTime, en.SuccessResponse, en.FailRate, en.MinRTT, en.MaxRTT, en.AvgRTT, en.AddInfo)
						if err != nil {
							// handle the "database is locked" error
							if strings.Contains(err.Error(), "database is locked") {
								time.Sleep(time.Millisecond * 100)
							} else {
								errChan <- err
								break
							}
						} else {
							break
						}
					}
				case "http":
					en := entry.(*RecordHTTPEntry)
					// Construct SQL query with the dynamic table name
					query := fmt.Sprintf("INSERT INTO %s (seq, status, response_code, response_phase, response_time, send_datetime, success_response, failure_rate, min_rtt, max_rtt, avg_rtt, additional_info) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", tableName)

					// Execute the query safely with placeholders for values
					for {
						_, err = ntdb.Exec(query, en.Seq, en.Status, en.ResponseCode, en.ResponsePhase, en.ResponseTime, en.SendDateTime, en.SuccessResponse, en.FailRate, en.MinRTT, en.MaxRTT, en.AvgRTT, en.AddInfo)
						if err != nil {
							// handle the "database is locked" error
							if strings.Contains(err.Error(), "database is locked") {
								time.Sleep(time.Millisecond * 100)
							} else {
								errChan <- err
								break
							}
						} else {
							break
						}
					}
				case "tcp":
					en := entry.(*RecordTCPEntry)
					// Construct SQL query with the dynamic table name
					query := fmt.Sprintf("INSERT INTO %s (seq, status, rtt, send_datetime, packetrecv, packetloss_rate, min_rtt, max_rtt, avg_rtt, additional_info) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", tableName)

					// Execute the query safely with placeholders for values
					for {
						_, err = ntdb.Exec(query, en.Seq, en.Status, en.RTT, en.SendDateTime, en.PacketRecv, en.PacketLossRate, en.MinRTT, en.MaxRTT, en.AvgRTT, en.AddInfo)
						if err != nil {
							// handle the "database is locked" error
							if strings.Contains(err.Error(), "database is locked") {
								time.Sleep(time.Millisecond * 100)
							} else {
								errChan <- err
								break
							}
						} else {
							break
						}
					}
				case "icmp":
					en := entry.(*RecordICMPEntry)
					// Construct SQL query with the dynamic table name
					query := fmt.Sprintf("INSERT INTO %s (seq, status, rtt, send_datetime, packetrecv, packetloss_rate, min_rtt, max_rtt, avg_rtt, additional_info) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", tableName)

					// Execute the query safely with placeholders for values
					for {
						_, err = ntdb.Exec(query, en.Seq, en.Status, en.RTT, en.SendDateTime, en.PacketRecv, en.PacketLossRate, en.MinRTT, en.MaxRTT, en.AvgRTT, en.AddInfo)
						if err != nil {
							// handle the "database is locked" error
							if strings.Contains(err.Error(), "database is locked") {
								time.Sleep(time.Millisecond * 100)
							} else {
								errChan <- err
								break
							}
						} else {
							break
						}
					}
				}
			}
		}
	}
}

// ReadHistoryTable retrieves all log entries and appends them to the provided *[]HistoryEntry
// func ReadHistoryTable(db *sql.DB, historyEntries *[]HistoryEntry) error {

// 	// initial []HistoryEntry
// 	*historyEntries = []HistoryEntry{}

// 	// Construct query dynamically
// 	query := "SELECT id, tablename, testtype, starttime, command, uuid, recorded FROM history;"

// 	// Execute the query
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return fmt.Errorf("error reading table %s: %v", "history", err)
// 	}
// 	defer rows.Close()

// 	// Iterate over rows and scan data into struct
// 	for rows.Next() {
// 		var entry HistoryEntry
// 		var recordedInt int // temporary variable to store the INT value of recorded

// 		// The rows.Scan() function in Go is used to map database query results into Go variables
// 		if err := rows.Scan(&entry.Id, &entry.TableName, &entry.TestType, &entry.StartTime, &entry.Command, &entry.UUID, &recordedInt); err != nil {
// 			return err
// 		}
// 		// update recorded
// 		if recordedInt == 1 {
// 			entry.Recorded = true
// 		} else {
// 			entry.Recorded = false
// 		}

// 		*historyEntries = append(*historyEntries, entry) // Modify the original slice
// 	}

// 	// Check for iteration errors
// 	if err := rows.Err(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// ReadHistoryTable retrieves all log entries and appends them to the provided *[]HistoryEntry
func ReadTableEntries(db *sql.DB, tableName string) (DbEntriesPointer *[]DbEntry, err error) {

	// initial DbEntries
	DbEntries := []DbEntry{}
	DbEntriesPointer = &DbEntries

	// read table based on table name or type
	if tableName == "history" {
		// Construct query dynamically
		query := "SELECT id, tablename, testtype, starttime, command, uuid, recorded FROM history;"

		// Execute the query
		rows, errIn := db.Query(query)
		if errIn != nil {
			err = fmt.Errorf("error reading table %s: %v", "history", errIn)
			return
		}
		defer rows.Close()

		// Iterate over rows and scan data into struct
		for rows.Next() {
			var entry HistoryEntry
			var recordedInt int // temporary variable to store the INT value of recorded

			// The rows.Scan() function in Go is used to map database query results into Go variables
			if errIn := rows.Scan(&entry.Id, &entry.TableName, &entry.TestType, &entry.StartTime, &entry.Command, &entry.UUID, &recordedInt); errIn != nil {
				err = errIn
				return
			}
			// update recorded
			if recordedInt == 1 {
				entry.Recorded = true
			} else {
				entry.Recorded = false
			}

			// append DbEntries
			DbEntries = append(DbEntries, &entry)
		}

		// Check for iteration errors
		if errIn := rows.Err(); errIn != nil {
			err = errIn
			return
		}

	} else {
		tableType := (strings.Split(tableName, "_"))[0]

		switch tableType {
		case "dns":
			// Construct query dynamically
			query := fmt.Sprintf("SELECT seq, status, dns_response, record, response_time, send_datetime, success_response, failure_rate, min_rtt, max_rtt, avg_rtt, additional_info FROM %s;", tableName)

			// Execute the query
			rows, errIn := db.Query(query)
			if errIn != nil {
				err = fmt.Errorf("error reading table %s: %v", tableName, errIn)
				return
			}
			defer rows.Close()

			// Iterate over rows and scan data into struct
			for rows.Next() {
				var entry RecordDNSEntry

				// The rows.Scan() function in Go is used to map database query results into Go variables
				if errIn := rows.Scan(&entry.Seq, &entry.Status, &entry.DnsResponse, &entry.DnsRecord, &entry.ResponseTime, &entry.SendDateTime, &entry.SuccessResponse, &entry.FailRate, &entry.MinRTT, &entry.MaxRTT, &entry.AvgRTT, &entry.AddInfo); errIn != nil {
					err = errIn
					return
				}

				// append DbEntries
				DbEntries = append(DbEntries, &entry)
			}

			// Check for iteration errors
			if errIn := rows.Err(); errIn != nil {
				err = errIn
				return
			}

		case "http":
			// Construct query dynamically
			query := fmt.Sprintf("SELECT seq, status, response_code, response_phase, response_time, send_datetime, success_response, failure_rate, min_rtt, max_rtt, avg_rtt, additional_info FROM %s;", tableName)

			// Execute the query
			rows, errIn := db.Query(query)
			if errIn != nil {
				err = fmt.Errorf("error reading table %s: %v", tableName, errIn)
				return
			}
			defer rows.Close()

			// Iterate over rows and scan data into struct
			for rows.Next() {
				var entry RecordHTTPEntry

				// The rows.Scan() function in Go is used to map database query results into Go variables
				if errIn := rows.Scan(&entry.Seq, &entry.Status, &entry.ResponseCode, &entry.ResponsePhase, &entry.ResponseTime, &entry.SendDateTime, &entry.SuccessResponse, &entry.FailRate, &entry.MinRTT, &entry.MaxRTT, &entry.AvgRTT, &entry.AddInfo); errIn != nil {
					err = errIn
					return
				}

				// append DbEntries
				DbEntries = append(DbEntries, &entry)
			}

			// Check for iteration errors
			if errIn := rows.Err(); errIn != nil {
				err = errIn
				return
			}

		case "tcp":
			// Construct query dynamically
			query := fmt.Sprintf("SELECT seq, status, rtt, send_datetime, packetrecv, packetloss_rate, min_rtt, max_rtt, avg_rtt, additional_info FROM %s;", tableName)

			// Execute the query
			rows, errIn := db.Query(query)
			if errIn != nil {
				err = fmt.Errorf("error reading table %s: %v", tableName, errIn)
				return
			}
			defer rows.Close()

			// Iterate over rows and scan data into struct
			for rows.Next() {
				var entry RecordTCPEntry

				// The rows.Scan() function in Go is used to map database query results into Go variables
				if errIn := rows.Scan(&entry.Seq, &entry.Status, &entry.RTT, &entry.SendDateTime, &entry.PacketRecv, &entry.PacketLossRate, &entry.MinRTT, &entry.MaxRTT, &entry.AvgRTT, &entry.AddInfo); errIn != nil {
					err = errIn
					return
				}

				// append DbEntries
				DbEntries = append(DbEntries, &entry)
			}

			// Check for iteration errors
			if errIn := rows.Err(); errIn != nil {
				err = errIn
				return
			}

		case "icmp":
			// Construct query dynamically
			query := fmt.Sprintf("SELECT seq, status, rtt, send_datetime, packetrecv, packetloss_rate, min_rtt, max_rtt, avg_rtt, additional_info FROM %s;", tableName)

			// Execute the query
			rows, errIn := db.Query(query)
			if errIn != nil {
				err = fmt.Errorf("error reading table %s: %v", tableName, errIn)
				return
			}
			defer rows.Close()

			// Iterate over rows and scan data into struct
			for rows.Next() {
				var entry RecordICMPEntry

				// The rows.Scan() function in Go is used to map database query results into Go variables
				if errIn := rows.Scan(&entry.Seq, &entry.Status, &entry.RTT, &entry.SendDateTime, &entry.PacketRecv, &entry.PacketLossRate, &entry.MinRTT, &entry.MaxRTT, &entry.AvgRTT, &entry.AddInfo); errIn != nil {
					err = errIn
					return
				}

				// append DbEntries
				DbEntries = append(DbEntries, &entry)
			}

			// Check for iteration errors
			if errIn := rows.Err(); errIn != nil {
				err = errIn
				return
			}

		}
	}

	return
}

// Func: Delete entry from "table" by "key" & "value"
func DeleteEntry(db *sql.DB, table, key, value string) error {
	// Construct the delete query dynamically (tableName must be validated)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?;", table, key)

	// Execute the delete statement
	for {
		_, err := db.Exec(query, value)
		if err != nil {
			// handle the "database is locked" error
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(time.Millisecond * 100)
			} else {
				return fmt.Errorf("error deleting entry from %s: %v", table, err)
			}
		} else {
			break
		}
	}

	//fmt.Printf("Successfully deleted entry with ID %d from table %s.\n", id, tableName)
	return nil
}

// Func: Delete table based on table name
func DeleteTable(db *sql.DB, tableName string) error {

	query := fmt.Sprintf("DROP TABLE IF EXISTS %q", tableName)
	for {
		_, err := db.Exec(query)
		if err != nil {
			// handle the "database is locked" error
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(time.Millisecond * 100)
			} else {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

// display History Entries in Console
func ShowHistoryTableConsole(historyEntries *[]HistoryEntry) {
	// Print the history entries
	fmt.Println("")
	fmt.Println("History Entries:")
	for _, entry := range *historyEntries {
		fmt.Printf("ID: %s, TableName: %s, TestType: %s, StartTime: %s, Command: %s, UUID: %s, Recorded: %v\n", entry.Id, entry.TableName, entry.TestType, entry.StartTime, entry.Command, entry.UUID, entry.Recorded)
	}
}

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
			ssuccess_response TEXT,
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
			packetloss_rate TEXT,
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
			packetloss_rate TEXT,
			min_rtt TEXT,
			max_rtt TEXT,
			avg_rtt TEXT,
			additional_info TEXT
		);`, testTableName)
	}

	for {
		_, err := db.Exec(query)
		if err != nil {
			// handle the "database is locked" error
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(time.Millisecond * 100)
			} else {
				return fmt.Errorf("failed to create table %s: %v", testTableName, err)
			}
		} else {
			break
		}
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

// func update the recording field of History Table
func UpdateFieldValue(db *sql.DB, table, searchKey, searchType, searchValue, updateKey, updateType, updateValue string) (err error) {

	// examine the update value
	var updateVal interface{}

	if updateType == "int" {
		updateVal, err = strconv.Atoi(updateValue)
		if err != nil {
			return
		}
	} else {
		updateVal = updateValue
	}

	// examine the search value
	var searchVal interface{}

	if searchType == "int" {
		searchVal, err = strconv.Atoi(searchValue)
		if err != nil {
			return
		}
	} else {
		searchVal = searchValue
	}

	// Construct the update query dynamically (tableName must be validated)

	query := fmt.Sprintf(`UPDATE %s SET %s = $1 WHERE %s = $2`, table, updateKey, searchKey)

	for {
		_, err := db.Exec(query, updateVal, searchVal)
		if err != nil {
			// handle the "database is locked" error
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(time.Millisecond * 100)
			} else {
				return fmt.Errorf("failed to update field: %w", err)
			}
		} else {
			break
		}
	}
	return nil
}

// func: *[]DbEntry -> *[]HistoryEntry
func ConvertDbEntriesToHistoryEntries(entries *[]DbEntry) (*[]HistoryEntry, error) {
	var historyEntries []HistoryEntry
	for _, entry := range *entries {
		if h, ok := entry.(*HistoryEntry); ok {
			historyEntries = append(historyEntries, *h)
		} else {
			return nil, fmt.Errorf("entry is not of type *HistoryEntry: %+v", entry)
		}
	}
	return &historyEntries, nil
}

// func: *[]DbEntry -> *[]RecordDNSEntry
func ConvertDbEntriesToRecordDNSEntries(entries *[]DbEntry) (*[]RecordDNSEntry, error) {
	var RecordDNSEntries []RecordDNSEntry
	for _, entry := range *entries {
		if r, ok := entry.(*RecordDNSEntry); ok {
			RecordDNSEntries = append(RecordDNSEntries, *r)
		} else {
			return nil, fmt.Errorf("entry is not of type *RecordDNSEntry: %+v", entry)
		}
	}
	return &RecordDNSEntries, nil
}

// func: *[]DbEntry -> *[]RecordHTTPEntry
func ConvertDbEntriesToRecordHTTPEntries(entries *[]DbEntry) (*[]RecordHTTPEntry, error) {
	var RecordHTTPEntries []RecordHTTPEntry
	for _, entry := range *entries {
		if r, ok := entry.(*RecordHTTPEntry); ok {
			RecordHTTPEntries = append(RecordHTTPEntries, *r)
		} else {
			return nil, fmt.Errorf("entry is not of type *RecordHTTPEntry: %+v", entry)
		}
	}
	return &RecordHTTPEntries, nil
}

// func: *[]DbEntry -> *[]RecordTCPEntry
func ConvertDbEntriesToRecordTCPEntries(entries *[]DbEntry) (*[]RecordTCPEntry, error) {
	var RecordTCPEntries []RecordTCPEntry
	for _, entry := range *entries {
		if r, ok := entry.(*RecordTCPEntry); ok {
			RecordTCPEntries = append(RecordTCPEntries, *r)
		} else {
			return nil, fmt.Errorf("entry is not of type *RecordTCPEntry: %+v", entry)
		}
	}
	return &RecordTCPEntries, nil
}
