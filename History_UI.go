package main

import (
	"database/sql"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func HistoryContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.Entry) *fyne.Container {

	// Return your History interface components here
	insertBtn := widget.NewButton("Insert Entry", func() {})
	refreshBtn := widget.NewButtonWithIcon("View Refresh", theme.ViewRefreshIcon(), func() {})

	deleteEntry := widget.NewEntry()
	deleteBtn := widget.NewButton("delete Entry", func() {})
	deleteContainer := container.New(layout.NewHBoxLayout(), deleteEntry, deleteBtn)

	btnContainer := container.New(layout.NewVBoxLayout(), insertBtn, refreshBtn, deleteContainer)

	// initoal entries slide
	historyEntries := []ntdb.HistoryEntry{}

	// insert Btn functions
	insertBtn.OnTapped = func() {
		he := ntdb.HistoryEntry{}

		Now := time.Now()
		he.TableName = "history"
		he.Date = Now.Format("2006-01-02")
		he.Time = Now.Format("15:04:05 MST")
		he.Type = "dns"
		he.Command = "nt -r dns 8.8.8.8 google.com"
		//he.Info = "No Error"

		// insert to entryChan
		entryChan <- &he

		err := historyRefresh(db, &historyEntries)
		if err != nil {
			logger.Println(err)
		}
	}

	// refresh table Btn
	refreshBtn.OnTapped = func() {
		err := historyRefresh(db, &historyEntries)
		if err != nil {
			logger.Println(err)
		}
	}

	// delete entry Btn
	deleteBtn.OnTapped = func() {
		id, _ := strconv.Atoi(deleteEntry.Text)
		err := ntdb.DeleteEntryByID(db, "history", id)
		if err != nil {
			logger.Println(err)
		}
		// refresh
		err = historyRefresh(db, &historyEntries)
		if err != nil {
			logger.Println(err)
		}
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(btnContainer, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), btnContainer, HistorySpaceHolder)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
