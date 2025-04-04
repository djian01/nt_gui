package main

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func HistoryContainer(a fyne.App, w fyne.Window, historyEntries *[]ntdb.HistoryEntry, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) *fyne.Container {

	// history table column: id, type, start time, command, btn: record, delete, replay

	// ** Refresh-Button Card **
	historyRefreshBtn := widget.NewButtonWithIcon("Rresh", theme.ViewRefreshIcon(), func() {})
	historyRefreshBtn.Importance = widget.HighImportance
	historyRefreshBtnContainerInner := container.NewGridWrap(fyne.NewSize(120, 30), historyRefreshBtn)
	historyRefreshBtnContainer := container.New(layout.NewBorderLayout(nil, nil, historyRefreshBtnContainerInner, nil), historyRefreshBtnContainerInner)
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
		err := historyRefresh(a, w, historyEntries, db, entryChan, errChan)
		if err != nil {
			logger.Println(err)
		}
	}

	// update the history table at the beginning
	err := historyRefresh(a, w, historyEntries, db, entryChan, errChan)
	if err != nil {
		logger.Println(err)
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyRefreshBtnBtncard, nil, nil, nil), historyRefreshBtnBtncard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
