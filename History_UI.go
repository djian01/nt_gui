package main

import (
	"database/sql"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

func HistoryContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error, selectedEntries *[]selectedEntry, selectAllCheckBox *widget.Check) *fyne.Container {

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
		err := historyRefresh(a, w, db, entryChan, errChan, filter, selectedEntries, selectAllCheckBox)
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
		err := historyRefresh(a, w, db, entryChan, errChan, "ALL", selectedEntries, selectAllCheckBox)
		if err != nil {
			logger.Println(err)
		}
	}

	//// *** select all check box func ***
	selectAllCheckBox.OnChanged = func(b bool) {

		// select operation
		selectOperation(db, errChan, selectedEntries, b)

		// fresh history table
		err := historyRefresh(a, w, db, entryChan, errChan, filter, selectedEntries, selectAllCheckBox)
		if err != nil {
			logger.Println(err)
		}
	}

	// delete-selected btn
	deleteSelectedBtn := widget.NewButtonWithIcon("Delete Selected", theme.DeleteIcon(), func() {
		// get history entries
		historyEntries := GetHistoryEntries(db, errChan)

		confirm := dialog.NewConfirm("Please Confirm", fmt.Sprintf("Do you want to delete %v history records?\n", len(*selectedEntries)), func(b bool) {
			if b {
				// record the deleted entry UUIDs
				deletedEntries := []selectedEntry{}

				// delete entry
				for _, s := range *selectedEntries {

					// record the deleted UUID
					deletedEntries = append(deletedEntries, selectedEntry{s.UUID, s.testType})

					// delete history entry
					err := ntdb.DeleteEntry(db, "history", "uuid", s.UUID)
					if err != nil {
						errChan <- err
						return
					}

					// check if the history entry is recorded
					recordFlag := false
					for _, he := range *historyEntries {
						if he.UUID == s.UUID && he.Recorded {
							recordFlag = true
							break
						}
					}

					// delete record table
					if recordFlag {
						err = ntdb.DeleteTable(db, fmt.Sprintf("%s_%s", s.testType, s.UUID))
						if err != nil {
							errChan <- err
							return
						}
					}
				}

				// remote the Selected Entry
				for _, r := range deletedEntries {
					// delete select entry
					DelSelectedEntry(selectedEntries, r)
				}

				// refresh table
				err := historyRefresh(a, w, db, entryChan, errChan, "ALL", selectedEntries, selectAllCheckBox)
				if err != nil {
					errChan <- err
				}
			}
		}, w)

		confirm.Show()

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

	//// *** body ***
	historyTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.historyTable = historyTableBody

	historyTableScroll := container.NewScroll(ntGlobal.historyTable)
	hisotryTableContainer := container.New(layout.NewBorderLayout(historyHeaderRow, nil, nil, nil), historyHeaderRow, historyTableScroll)

	// ** Table Card **
	historyTableCard := widget.NewCard("", "", hisotryTableContainer)

	// update the history table at the beginning
	err := historyRefresh(a, w, db, entryChan, errChan, "ALL", selectedEntries, selectAllCheckBox)
	if err != nil {
		logger.Println(err)
	}

	// ** Main Container **
	HistorySpaceHolder := widget.NewLabel("    ")
	HistoryMainContainerIner := container.New(layout.NewBorderLayout(historyActionCard, nil, nil, nil), historyActionCard, historyTableCard)
	HistoryMainContainerOuter := container.New(layout.NewBorderLayout(HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder, HistorySpaceHolder), HistorySpaceHolder, HistoryMainContainerIner)

	return HistoryMainContainerOuter // Temporary empty container, replace with your actual UI
}
