package main

import (
	"database/sql"
	"image/color"

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
	d.TestType.Length = 50
	d.TestType.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	d.StartTime.Label = "StartTime"
	d.StartTime.Length = 190
	d.StartTime.Object = widget.NewLabel("--")

	d.Command.Label = "NTCommand"
	d.Command.Length = 250
	d.Command.Object = widget.NewLabel("--")

	d.ReplayBtn = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {})
	d.ReplayBtn.Importance = widget.HighImportance

	d.RecordBtn = widget.NewButtonWithIcon("", theme.FileIcon(), func() {})
	d.RecordBtn.Importance = widget.WarningImportance
	d.RecordBtn.Disable()

	d.DeleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.DeleteBtn.Importance = widget.DangerImportance

	d.Action.Label = "Action"
	d.Action.Length = 110
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(3), d.ReplayBtn, d.RecordBtn, d.ReplayBtn)

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
	d.StartTime.Object.(*widget.Label).Text = h.DateTime
	d.StartTime.Object.(*widget.Label).Refresh()

	// Command
	d.Command.Object.(*widget.Label).Text = h.Command
	d.Command.Object.(*widget.Label).Refresh()

	// Action
	if h.Recorded {
		d.RecordBtn.Enable()
	}

}

func historyRefresh(db *sql.DB, historyEntries *[]ntdb.HistoryEntry) error {

	err := ntdb.ReadHistoryTable(db, historyEntries)
	if err != nil {
		return err
	}
	// show all the history table in console
	ntdb.ShowHistoryTableConsole(historyEntries)

	return nil
}
