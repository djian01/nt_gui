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

	// ** Action Card **

	// **** filter btn

	// Create resource from SVG file
	filterIcon := theme.NewThemedResource(resourceFilterSvg)

	// filter select
	filter := "--"
	historyFilterSelect := widget.NewSelect([]string{"ALL", "ICMP", "TCP", "HTTP", "DNS"}, func(s string) {
		switch s {
		case "ICMP":
			filter = "icmp"
		case "TCP":
			filter = "tcp"
		case "HTTP":
			filter = "http"
		case "DNS":
			filter = "dns"
		default:
			filter = "--"
		}
	})

	historyFilterSelect.Selected = "ALL"

	historyFilterSelectContainerInner := container.NewGridWrap(fyne.NewSize(100, 30), historyFilterSelect)

	// filter btn
	historyFilterBtn := widget.NewButtonWithIcon("Refresh", filterIcon, func() {
		err := historyRefresh(a, w, historyEntries, db, entryChan, errChan, filter)
		if err != nil {
			errChan <- err
			return
		}
	})
	historyFilterBtn.Importance = widget.HighImportance
	historyFilterBtnContainerInner := container.NewGridWrap(fyne.NewSize(120, 30), historyFilterBtn)

	// filter container
	historyFilterContainer := container.New(layout.NewHBoxLayout(), historyFilterSelectContainerInner, historyFilterBtnContainerInner)

	// delete selected btn
	deleteSelectedBtn := widget.NewButtonWithIcon("Delete Selected", theme.DeleteIcon(), func() {})
	deleteSelectedBtn.Importance = widget.DangerImportance
	deleteSelectedBtnContainerInner := container.NewGridWrap(fyne.NewSize(180, 30), deleteSelectedBtn)

	historyActionContainer := container.New(layout.NewBorderLayout(nil, nil, deleteSelectedBtnContainerInner, historyFilterContainer), deleteSelectedBtnContainerInner, historyFilterContainer)
	historyActionCard := widget.NewCard("", "", historyActionContainer)

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

	// update the history table at the beginning
	err := historyRefresh(a, w, historyEntries, db, entryChan, errChan, "--")
	if err != nil {
		logger.Println(err)
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyActionCard, nil, nil, nil), historyActionCard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
