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

func HistoryContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry) *fyne.Container {

	// history table column: id, type, start time, command, btn: record, delete, replay

	// ** Refresh-Button Card **
	historyRefreshBtn := widget.NewButtonWithIcon("Rresh", theme.ViewRefreshIcon(), func() {})
	historyRefreshBtn.Importance = widget.HighImportance
	historyRefreshBtnContainer := container.New(layout.NewBorderLayout(nil, nil, historyRefreshBtn, nil), historyRefreshBtn)
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

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyRefreshBtnBtncard, nil, nil, nil), historyRefreshBtnBtncard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
