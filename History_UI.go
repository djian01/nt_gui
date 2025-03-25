package main

import (
	"database/sql"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func HistoryContainer(a fyne.App, w fyne.Window, historyEntries *[]ntdb.HistoryEntry, db *sql.DB, entryChan chan ntdb.DbEntry) *fyne.Container {

	// Initial History Entries Slice
	//historyEntries := []ntdb.HistoryEntry{}

	//// test code start
	insertBtn := widget.NewButton("Insert Entry", func() {})
	// insert Btn functions
	insertBtn.OnTapped = func() {
		he := ntdb.HistoryEntry{}

		Now := time.Now()
		he.TableName = "history"
		he.StartTime = Now.Format("2006-01-02 15:04:05 MST")
		he.TestType = "dns"
		he.Command = "nt -r dns 8.8.8.8 google.com"
		he.UUID = GenerateShortUUID()
		he.Recorded = false

		// insert to entryChan
		entryChan <- &he
	}

	//// test code ends

	// history table column: id, type, start time, command, btn: record, delete, replay

	// ** Refresh-Button Card **
	historyRefreshBtn := widget.NewButtonWithIcon("Rresh", theme.ViewRefreshIcon(), func() {})
	historyRefreshBtn.Importance = widget.HighImportance
	historyRefreshBtnContainerInner := container.NewGridWrap(fyne.NewSize(120, 30), historyRefreshBtn)
	historyRefreshBtnContainer := container.New(layout.NewBorderLayout(nil, nil, historyRefreshBtnContainerInner, nil), historyRefreshBtnContainerInner, insertBtn)
	historyRefreshBtnBtncard := widget.NewCard("", "", historyRefreshBtnContainer)

	// ** Table Container **
	historyHeader := historyGUIRow{}
	historyHeader.Initial()
	historyHeaderRow := historyHeader.GenerateHeaderRow()

	historyTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.historyTable = historyTableBody

	historyTableScroll := container.NewScroll(ntGlobal.historyTable)
	hisotryTableContainer := container.New(layout.NewBorderLayout(historyHeaderRow, nil, nil, nil), historyHeaderRow, historyTableScroll)

	// ** Table Card **
	historyTableCard := widget.NewCard("", "", hisotryTableContainer)

	// btn functions:
	//// history refresh btn
	historyRefreshBtn.OnTapped = func() {
		err := historyRefresh(a, w, historyEntries, db, entryChan)
		if err != nil {
			logger.Println(err)
		}
	}

	// update the history table at the beginning
	err := historyRefresh(a, w, historyEntries, db, entryChan)
	if err != nil {
		logger.Println(err)
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyRefreshBtnBtncard, nil, nil, nil), historyRefreshBtnBtncard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
