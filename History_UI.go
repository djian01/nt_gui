package main

import (
	"database/sql"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func HistoryContainer(a fyne.App, w fyne.Window, historyEntries *[]ntdb.HistoryEntry, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error, selectedEntries *[]selectedEntry, selectAllCheckBox *widget.Check) *fyne.Container {

	// history table column: selected id, type, start time, command, btn: record, delete, replay

	// ** Action Card **

	// Create resource from SVG file
	filterIcon := theme.NewThemedResource(resourceFilterSvg)

	// filter select
	filter := "ALL"
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
			filter = "ALL"
		}
	})

	historyFilterSelect.Selected = "ALL"

	historyFilterSelectContainerInner := container.NewGridWrap(fyne.NewSize(100, 30), historyFilterSelect)

	// filter btn
	historyFilterBtn := widget.NewButtonWithIcon("Filter", filterIcon, func() {
		err := historyRefresh(a, w, historyEntries, db, entryChan, errChan, filter, selectedEntries, selectAllCheckBox)
		if err != nil {
			errChan <- err
			return
		}

	})
	historyFilterBtn.Importance = widget.HighImportance
	historyFilterBtnContainerInner := container.NewGridWrap(fyne.NewSize(120, 30), historyFilterBtn)

	// filter container
	historyFilterContainer := container.New(layout.NewHBoxLayout(), historyFilterSelectContainerInner, historyFilterBtnContainerInner)

	// select all onChange func
	selectAllCheckBox.OnChanged = func(b bool) {
		err := historyRefresh(a, w, historyEntries, db, entryChan, errChan, "ALL", selectedEntries, selectAllCheckBox)
		if err != nil {
			logger.Println(err)
		}
	}

	// delete-selected btn
	deleteSelectedBtn := widget.NewButtonWithIcon("Delete Selected", theme.DeleteIcon(), func() {
		for _, s := range *selectedEntries {
			fmt.Println(s)
		}
		fmt.Println("===================")
	})

	deleteSelectedBtn.Importance = widget.DangerImportance
	deleteSelectedBtnContainerInner := container.NewGridWrap(fyne.NewSize(180, 30), deleteSelectedBtn)

	historyActionContainer := container.New(layout.NewBorderLayout(nil, nil, deleteSelectedBtnContainerInner, historyFilterContainer), deleteSelectedBtnContainerInner, historyFilterContainer)
	historyActionCard := widget.NewCard("", "", historyActionContainer)

	// ** Table Container **

	//// *** header ***
	historyHeader := historyGUIRow{}
	historyHeader.Initial()
	historyHeaderRow := historyHeader.GenerateHeaderRow(selectAllCheckBox)

	selectAllCheckBox.OnChanged = func(b bool) {

	}

	//// *** body ***
	historyTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.historyTable = historyTableBody

	historyTableScroll := container.NewScroll(ntGlobal.historyTable)
	hisotryTableContainer := container.New(layout.NewBorderLayout(historyHeaderRow, nil, nil, nil), historyHeaderRow, historyTableScroll)

	// ** Table Card **
	historyTableCard := widget.NewCard("", "", hisotryTableContainer)

	// update the history table at the beginning
	err := historyRefresh(a, w, historyEntries, db, entryChan, errChan, "ALL", selectedEntries, selectAllCheckBox)
	if err != nil {
		logger.Println(err)
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyActionCard, nil, nil, nil), historyActionCard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}

// func: Add selected Entry
func AddSelectedEntry(selectedEntries *[]selectedEntry, selected selectedEntry) {
	for _, s := range *selectedEntries {
		if s.UUID == selected.UUID {
			return // UUID already exists
		}
	}
	*selectedEntries = append(*selectedEntries, selected)
}

// func: Delete a selected Entry
func DelSelectedEntry(selectedUUIDs *[]selectedEntry, selected selectedEntry) {
	for i, s := range *selectedUUIDs {
		if s.UUID == selected.UUID {
			if i == len(*selectedUUIDs)-1 {
				// UUID is the last element
				*selectedUUIDs = (*selectedUUIDs)[:i]
			} else {
				// UUID is not the last element
				*selectedUUIDs = append((*selectedUUIDs)[:i], (*selectedUUIDs)[i+1:]...)
			}
			return
		}
	}
}

// func: ExtryExist
func EntryExist(selectedEntries *[]selectedEntry, UUID string) bool {

	for _, s := range *selectedEntries {
		if s.UUID == UUID {
			return true
		}
	}
	return false
}
