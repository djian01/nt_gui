//go:build ignore
// +build ignore

package ntdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	_ "github.com/xeodou/go-sqlcipher"
)

func DBOpenS(dbFile string) (*sql.DB, error) {

	key := "123456"
	passPhase := fmt.Sprintf("%s?_key=%s", dbFile, key)

	// check "dbFile" in the same folder as the executable
	// if db file not exist, os.Stat(dbFile) return error. And os.IsNotExist(err) returns true if err exist
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		// fmt.Println("Database file not found, creating new database...")
		err := createDatabaseS(dbFile, passPhase)
		if err != nil {
			return nil, errors.New("failed to create database")
		}
		// fmt.Println("Database created successfully!")
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", passPhase)
	if err != nil {
		return nil, errors.New("failed to open database")
	}

	return db, nil
}

// createDatabase creates a new SQLite database file and initializes it with a default table
func createDatabaseS(dbFile, passPhase string) error {
	// Create an empty database file
	file, err := os.Create(dbFile)
	if err != nil {
		return err
	}
	file.Close() // Close immediately since SQLite will manage it

	// Open SQLite connection
	db, err := sql.Open("sqlite3", passPhase)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create a default table (example: users)
	return createHistoryTable(db)
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
			success_response INTEGER,
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
			successresponse INTEGER,
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
			packetrecv INTEGER,
			packetloss INTEGER,
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
			packetrecv INTEGER,
			packetloss INTEGER,
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
