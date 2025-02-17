package ntdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

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

// createDatabase creates a new SQLite database file and initializes it with a default table
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

	// Create a default table (example: users)
	return createHistoryTable(db)
}

// createDefaultTable creates that a "users" table exists in the database
func createHistoryTable(db *sql.DB) error {

	// default table name is "history"
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		time TEXT NOT NULL,
		type TEXT NOT NULL,
		command TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

// InsertEntry inserts a log entry into the "history" table
func InsertEntry(ntdb *sql.DB, entryChan <-chan Entry) error {

	// initial err
	var err error = nil

	// read from channel
	for entry := range entryChan {
		tableName := entry.GetTableName()
		switch tableName {
		// history table
		case "history":
			he := entry.(*HistoryEntry)
			// Construct SQL query with the dynamic table name, default table name is "history"
			query := `INSERT INTO history (date, time, type, command) VALUES (?, ?, ?, ?);`

			// Execute the query safely with placeholders for values
			_, err = ntdb.Exec(query, he.Date, he.Time, he.Type, he.Command)
		}
	}
	return err
}

// ReadHistoryTable retrieves all log entries and appends them to the provided *[]HistoryEntry
func ReadHistoryTable(db *sql.DB, historyEntries *[]HistoryEntry) error {

	// initial []HistoryEntry
	*historyEntries = []HistoryEntry{}

	// Construct query dynamically
	query := "SELECT id, type, date, time, command FROM history;"

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error reading table %s: %v", "history", err)
	}
	defer rows.Close()

	// Iterate over rows and scan data into struct
	for rows.Next() {
		var entry HistoryEntry

		// The rows.Scan() function in Go is used to map database query results into Go variables
		if err := rows.Scan(&entry.Id, &entry.Type, &entry.Date, &entry.Time, &entry.Command); err != nil {
			return err
		}
		*historyEntries = append(*historyEntries, entry) // Modify the original slice
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// DeleteEntryByID deletes an entry from the given table using its ID
func DeleteEntryByID(db *sql.DB, tableName string, id int) error {
	// Construct the delete query dynamically (tableName must be validated)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?;", tableName)

	// Execute the delete statement
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting entry from %s: %v", tableName, err)
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error retrieving affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no entry found with ID %d in table %s", id, tableName)
	}

	fmt.Printf("Successfully deleted entry with ID %d from table %s.\n", id, tableName)
	return nil
}

// display History Entries in Console
func ShowHistoryTableConsole(historyEntries *[]HistoryEntry) {
	// Print the history entries
	fmt.Println("")
	fmt.Println("History Entries:")
	for _, entry := range *historyEntries {
		fmt.Printf("ID: %s, Type: %s, Date: %s, Time: %s, Command: %s\n", entry.Id, entry.Type, entry.Date, entry.Time, entry.Command)
	}
}
