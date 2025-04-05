package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

// ******* struct historyGUIRow ********

type historyGUIRow struct {
	Index           pingCell
	TestType        pingCell
	StartTime       pingCell
	Command         pingCell // fixed
	Action          pingCell
	ShowRecordBtn   *widget.Button
	DeleteBtn       *widget.Button
	ReplayBtn       *widget.Button
	historyTableRow *fyne.Container
}

func (d *historyGUIRow) Initial() {

	d.Index.Label = "Index"
	d.Index.Length = 50
	d.Index.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	d.TestType.Label = "TestType"
	d.TestType.Length = 120
	d.TestType.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	d.StartTime.Label = "StartTime"
	d.StartTime.Length = 250
	d.StartTime.Object = widget.NewLabel("--")

	d.Command.Label = "NTCommand"
	d.Command.Length = 450
	d.Command.Object = widget.NewLabel("--")

	d.ReplayBtn = widget.NewButtonWithIcon("Re-Run", theme.MediaReplayIcon(), func() {})
	d.ReplayBtn.Importance = widget.HighImportance

	d.ShowRecordBtn = widget.NewButtonWithIcon("Show Details", theme.FileIcon(), func() {})
	d.ShowRecordBtn.Importance = widget.WarningImportance
	d.ShowRecordBtn.Disable()

	d.DeleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.DeleteBtn.Importance = widget.DangerImportance

	d.Action.Label = "Action"
	d.Action.Length = 300
	d.Action.Object = container.New(layout.NewBorderLayout(nil, nil, nil, d.DeleteBtn), container.New(layout.NewGridLayoutWithColumns(2), d.ReplayBtn, d.ShowRecordBtn), d.DeleteBtn)

	// table row
	row := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), container.NewCenter(d.Index.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.TestType.Length), 30), container.NewCenter(d.TestType.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), container.NewCenter(d.StartTime.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Command.Length), 30), container.NewCenter(d.Command.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), d.Action.Object),
		GUIVerticalSeparator(),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 2)) // Adjust width & thickness

	d.historyTableRow = container.New(layout.NewVBoxLayout(),
		row,
		thickLine,
	)
}

func (d *historyGUIRow) GenerateHeaderRow() *fyne.Container {

	// table row
	header := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), widget.NewLabelWithStyle(d.Index.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.TestType.Length), 30), widget.NewLabelWithStyle(d.TestType.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), widget.NewLabelWithStyle(d.StartTime.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Command.Length), 30), widget.NewLabelWithStyle(d.Command.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), widget.NewLabelWithStyle(d.Action.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 3)) // Adjust width & thickness

	headerRow := container.New(layout.NewVBoxLayout(),
		header,
		thickLine,
	)

	return headerRow
}

func (d *historyGUIRow) UpdateRow(h *ntdb.HistoryEntry) {

	// Index
	d.Index.Object.(*widget.Label).Text = h.Id
	d.Index.Object.(*widget.Label).Refresh()

	// Test Type
	d.TestType.Object.(*widget.Label).Text = h.TestType
	d.TestType.Object.(*widget.Label).Refresh()

	// Start Time
	d.StartTime.Object.(*widget.Label).Text = h.StartTime.Format("2006-01-02 15:04:05 MST")
	d.StartTime.Object.(*widget.Label).Refresh()

	// Command
	d.Command.Object.(*widget.Label).Text = h.Command
	d.Command.Object.(*widget.Label).Refresh()

	// Action
	if h.Recorded {
		d.ShowRecordBtn.Enable()
	}

}

// ******* struct historyObject ********
type historyObject struct {
	historyEntry *ntdb.HistoryEntry
	historyGUI   historyGUIRow
}

// Func: add history row
func historyAddRow(a fyne.App, w fyne.Window, he *ntdb.HistoryEntry, hs *[]ntdb.HistoryEntry, historyTableBody *fyne.Container, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) {

	// create history object
	ho := historyObject{}
	ho.historyEntry = he
	ho.historyGUI.Initial()
	ho.historyGUI.UpdateRow(he)

	// check if the UUID is NOT in the registered running test, add the record in the history table
	if !existingTestCheck(&testRegister, he.UUID) {
		// update table body
		historyTableBody.Add(ho.historyGUI.historyTableRow)
		historyTableBody.Refresh()
	}

	// update replay btn
	ho.historyGUI.ReplayBtn.OnTapped = func() {
		// fmt.Println(he.historyEntry.UUID)
		// re-launch a new go routine for DnsAddPingRow with the same InputVar

		// generate iv
		recording, iv, _ := NtCmd2Iv(he.Command)

		switch he.TestType {
		case "dns":
			go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, recording, db, entryChan, errChan)
		case "http":
		case "tcp":
		case "icmp":
		}
	}

	// update show record details btn
	if he.Recorded {
		ho.historyGUI.ShowRecordBtn.Enable()
	} else {
		ho.historyGUI.ShowRecordBtn.Disable()
	}

	ho.historyGUI.ShowRecordBtn.OnTapped = func() {

		// dbTestEntries
		dbTestEntries, err := ntdb.ReadTestTableEntries(db, fmt.Sprintf("%s_%s", he.TestType, he.UUID))
		if err != nil {
			errChan <- err
			return
		}

		// generate Summary
		sumData, err := DbTestEntry2SummaryData(*he, (*dbTestEntries)[0], (*dbTestEntries)[len(*dbTestEntries)-1])
		if err != nil {
			errChan <- err
			return
		}
		// test
		fmt.Println(sumData)

		// []ntchart.chartPoint
		chartData := ntchart.ConvertFromDbToCheckpoint(dbTestEntries)

		// test
		fmt.Println(len(*chartData))

		// For Exporting CSV
		switch he.TestType {
		case "dns":
			entries, err := ntdb.ConvertDbTestEntriesToRecordDNSEntries(dbTestEntries)
			if err != nil {
				errChan <- err
				return
			}

			// test
			fmt.Println((*entries)[0])

		case "http":
		case "tcp":
		case "icmp":
		}

	}

	// update delete btn func
	ho.historyGUI.DeleteBtn.OnTapped = func() {

		confirm := dialog.NewConfirm("Please Confirm", fmt.Sprintf("Do you want to delete the history record of \n \"%s\" ?", he.Command), func(b bool) {
			if b {
				// delete entry
				err := ntdb.DeleteEntry(db, "history", "uuid", ho.historyEntry.UUID)
				if err != nil {
					errChan <- err
					return
				}

				// delete record table
				if he.Recorded {
					err = ntdb.DeleteTable(db, fmt.Sprintf("%s_%s", he.TestType, he.UUID))
					if err != nil {
						errChan <- err
						return
					}
				}

				// refresh table
				err = historyRefresh(a, w, hs, db, entryChan, errChan)
				if err != nil {
					errChan <- err
				}
			}
		}, w)

		confirm.Show()
	}
}

// Func: history table refresh
func historyRefresh(a fyne.App, w fyne.Window, historyEntries *[]ntdb.HistoryEntry, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) error {

	// clean up all the items in ntGlobal.historyTable
	ntGlobal.historyTable.Objects = nil // Remove all child objects

	// read the DB and obtain the historyEntries
	DbEntries, err := ntdb.ReadTableEntries(db, "history")
	if err != nil {
		return err
	}

	// convert *[]DbEntry -> *[]ntdb.HistoryEntry
	historyEntries, err = ntdb.ConvertDbEntriesToHistoryEntries(DbEntries)
	if err != nil {
		return err
	}

	// update history table body
	if len(*historyEntries) == 0 {
		return nil
	} else {
		for i := 0; i < len(*historyEntries); i++ {
			go historyAddRow(a, w, &(*historyEntries)[i], historyEntries, ntGlobal.historyTable, db, entryChan, errChan)
			// add some delays between each row to let the table sort by Id sequence
			time.Sleep(5 * time.Millisecond)
		}
	}

	return nil
}
