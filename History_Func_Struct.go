package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

// ******* struct historyGUIRow ********

type historyGUIRow struct {
	Index           pingCell
	TestType        pingCell
	StartTime       pingCell
	Command         pingCell // fixed
	Action          pingCell
	RecordBtn       *widget.Button
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

	d.RecordBtn = widget.NewButtonWithIcon("Show Details", theme.FileIcon(), func() {})
	d.RecordBtn.Importance = widget.WarningImportance
	d.RecordBtn.Disable()

	d.DeleteBtn = widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {})
	d.DeleteBtn.Importance = widget.DangerImportance

	d.Action.Label = "Action"
	d.Action.Length = 380
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(3), d.ReplayBtn, d.RecordBtn, d.DeleteBtn)

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
	d.StartTime.Object.(*widget.Label).Text = h.StartTime
	d.StartTime.Object.(*widget.Label).Refresh()

	// Command
	d.Command.Object.(*widget.Label).Text = h.Command
	d.Command.Object.(*widget.Label).Refresh()

	// Action
	if h.Recorded {
		d.RecordBtn.Enable()
	}

}

// ******* struct historyObject ********
type historyObject struct {
	historyEntry *ntdb.HistoryEntry
	historyGUI   historyGUIRow
}

// Func: add history row
func historyAddRow(a fyne.App, h *ntdb.HistoryEntry, hs *[]ntdb.HistoryEntry, historyTableBody *fyne.Container, db *sql.DB) {

	// create history object
	he := historyObject{}
	he.historyEntry = h
	he.historyGUI.Initial()
	he.historyGUI.UpdateRow(h)

	// check if the UUID is NOT in the registered running test, add the record in the history table
	if !existingTestCheck(&testRegister, h.UUID) {
		// update table body
		historyTableBody.Add(he.historyGUI.historyTableRow)
		historyTableBody.Refresh()
	}

	// update record btn
	if h.Recorded {
		he.historyGUI.RecordBtn.Enable()
	} else {
		he.historyGUI.RecordBtn.Disable()
	}

	he.historyGUI.RecordBtn.OnTapped = func() {

	}

	// update replay btn
	he.historyGUI.ReplayBtn.OnTapped = func() {
		fmt.Println(he.historyEntry.UUID)
	}

	// update delete btn
	he.historyGUI.DeleteBtn.OnTapped = func() {
		// delete entry
		err := ntdb.DeleteEntry(db, "history", "uuid", he.historyEntry.UUID)
		if err != nil {
			logger.Println(err)
		}
		// refresh table
		err = historyRefresh(a, db, hs)
		if err != nil {
			logger.Println(err)
		}
	}
}

// Func: history table refresh
func historyRefresh(a fyne.App, db *sql.DB, historyEntries *[]ntdb.HistoryEntry) error {

	// clean up all the items in ntGlobal.historyTable
	ntGlobal.historyTable.Objects = nil // Remove all child objects

	// read the DB and obtain the historyEntries
	err := ntdb.ReadHistoryTable(db, historyEntries)
	if err != nil {
		return err
	}

	// update history table body
	if len(*historyEntries) == 0 {
		return nil
	} else {
		for i := 0; i < len(*historyEntries); i++ {
			go historyAddRow(a, &(*historyEntries)[i], historyEntries, ntGlobal.historyTable, db)
			// add some delays between each row to let the table sort by Id sequence
			time.Sleep(5 * time.Millisecond)
		}
	}

	return nil
}
