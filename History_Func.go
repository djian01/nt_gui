package main

import (
	"database/sql"

	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func historyRefresh(db *sql.DB, historyEntries *[]ntdb.HistoryEntry) error {

	err := ntdb.ReadHistoryTable(db, historyEntries)
	if err != nil {
		return err
	}
	// show all the history table in console
	ntdb.ShowHistoryTableConsole(historyEntries)

	return nil
}
