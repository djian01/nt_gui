//go:build ignore
// +build ignore

package ntdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

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
