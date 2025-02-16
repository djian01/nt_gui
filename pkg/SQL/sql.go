package sql

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type InsertRequest struct {
	TableName string
	Headers   []string
	Data      []string
}

func main() {
	dbFile := "data.db"
	// check "data.db" in the same folder as the executable
	// if db file not exist, os.Stat(dbFile) return error. And os.IsNotExist(err) returns true if err exist
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Println("Database file not found, creating new database...")
		err := createDatabase(dbFile)
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
		fmt.Println("Database created successfully!")
	} else {
		fmt.Println("Database already exists.")
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a channel for writing to the DB
	writeChan := make(chan InsertRequest, 10)

	// Start the writer goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go dbWriter(db, writeChan, &wg)

	// Read CSV file and send data to the channel
	err = readCSVAndSave("data.csv", writeChan)
	if err != nil {
		log.Fatal(err)
	}

	// Close the write channel and wait for writer to finish
	close(writeChan)
	wg.Wait()

	fmt.Println("Data successfully saved to database!")

	// Read data from the database
	tableName := "data" // Adjust as needed
	fmt.Printf("\nReading data from table: %s\n", tableName)
	err = readFromTable(db, tableName)
	if err != nil {
		log.Fatal(err)
	}
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
	return ensureDefaultTable(db)
}

// ensureDefaultTable ensures that a "users" table exists in the database
func ensureDefaultTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

// dbWriter listens for InsertRequests and writes to the database
func dbWriter(db *sql.DB, writeChan chan InsertRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range writeChan {
		// Ensure the table exists
		createTableIfNotExists(db, req.TableName, req.Headers)

		// Insert data into table
		insertIntoTable(db, req.TableName, req.Headers, req.Data)
	}
}

// createTableIfNotExists dynamically creates a table based on headers
func createTableIfNotExists(db *sql.DB, tableName string, headers []string) {
	fields := []string{}
	for _, h := range headers {
		fields = append(fields, fmt.Sprintf("%s TEXT", h))
	}
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(fields, ", "))

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating table %s: %v", tableName, err)
	}
}

// insertIntoTable inserts data into the specified table
func insertIntoTable(db *sql.DB, tableName string, headers, data []string) {
	placeholders := strings.Repeat("?,", len(headers))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(headers, ", "), placeholders)

	_, err := db.Exec(query, convertToInterfaceSlice(data)...)
	if err != nil {
		log.Printf("Error inserting into %s: %v", tableName, err)
	}
}

// convertToInterfaceSlice converts a string slice to an interface{} slice for Exec()
func convertToInterfaceSlice(data []string) []interface{} {
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = v
	}
	return result
}

// readCSVAndSave reads a CSV file and sends data to the write channel
func readCSVAndSave(csvFile string, writeChan chan InsertRequest) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file must have at least a header row and one data row")
	}

	headers := records[0]                            // First row is the table header
	tableName := strings.TrimSuffix(csvFile, ".csv") // Use filename as table name

	// Send each row as an InsertRequest
	for _, row := range records[1:] {
		writeChan <- InsertRequest{TableName: tableName, Headers: headers, Data: row}
	}

	return nil
}

// readFromTable reads all rows from the given table and prints them
func readFromTable(db *sql.DB, tableName string) error {
	// Get column names
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error fetching table schema: %v", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		columns = append(columns, name)
	}

	if len(columns) == 0 {
		return fmt.Errorf("table %s does not exist or has no columns", tableName)
	}

	// Read table data
	query = fmt.Sprintf("SELECT * FROM %s;", tableName)
	rows, err = db.Query(query)
	if err != nil {
		return fmt.Errorf("error reading table: %v", err)
	}
	defer rows.Close()

	fmt.Println(strings.Join(columns, " | ")) // Print headers

	// Read and print rows
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		// Convert []interface{} to []string
		stringValues := make([]string, len(values))

		for i, val := range values {
			if val == nil {
				stringValues[i] = "NULL"
			} else {
				stringValues[i] = fmt.Sprintf("%v", val)
			}
		}

		fmt.Println(strings.Join(stringValues, " | "))
	}

	return nil
}
