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
	"github.com/djian01/nt/pkg/ntPinger"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	ntdb "github.com/djian01/nt_gui/pkg/ntdb"
)

// ******* struct historyGUIRow ********

type historyGUIRow struct {
	Selected        pingCell
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

	d.Selected.Label = "Selected"
	d.Selected.Length = 50
	d.Selected.Object = widget.NewCheck("", func(b bool) {})

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
		container.NewGridWrap(fyne.NewSize(float32(d.Selected.Length), 30), container.NewCenter(d.Selected.Object)),
		GUIVerticalSeparator(),
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

func (d *historyGUIRow) GenerateHeaderRow(selectAllCheckBox *widget.Check) *fyne.Container {

	// selected header
	selectAllContainer := container.New(layout.NewCenterLayout(), selectAllCheckBox)

	// table row
	header := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(float32(d.Selected.Length), 30), selectAllContainer),
		GUIVerticalSeparator(),
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
	d.Command.Object.(*widget.Label).Text = TruncateString(h.Command, 70)
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

// ******* struct historyObject ********
type selectedEntry struct {
	UUID     string
	testType string
}

// Func: add history row
func historyAddRow(a fyne.App, w fyne.Window, he *ntdb.HistoryEntry, historyTableBody *fyne.Container, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error, selectedEntries *[]selectedEntry, displayObjects *[]ntdb.HistoryEntry) {

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

		// generate iv
		recording, iv, _ := NtCmd2Iv(he.Command)

		switch he.TestType {
		case "dns":
			go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, recording, db, entryChan, errChan)
		case "http":
			go HttpAddPingRow(a, &ntGlobal.httpIndex, &iv, ntGlobal.httpTable, recording, db, entryChan, errChan)
		case "tcp":
			go TcpAddPingRow(a, &ntGlobal.tcpIndex, &iv, ntGlobal.tcpTable, recording, db, entryChan, errChan)
		case "icmp":
			go IcmpAddPingRow(a, &ntGlobal.icmpIndex, &iv, ntGlobal.icmpTable, recording, db, entryChan, errChan)
		}
	}

	// update select check if the current entry is already selected
	if EntryExist(selectedEntries, he.UUID) {
		(ho.historyGUI.Selected.Object.(*widget.Check)).SetChecked(true)
	} else {
		(ho.historyGUI.Selected.Object.(*widget.Check)).SetChecked(false)
	}

	// set select check func
	(ho.historyGUI.Selected.Object.(*widget.Check)).OnChanged = func(b bool) {
		// uncheck
		if !b {
			DelSelectedEntry(selectedEntries, selectedEntry{UUID: he.UUID, testType: he.TestType})
			// check
		} else {
			AddSelectedEntry(selectedEntries, selectedEntry{UUID: he.UUID, testType: he.TestType})
		}
	}

	// update show record details btn
	if he.Recorded {
		ho.historyGUI.ShowRecordBtn.Enable()
	} else {
		ho.historyGUI.ShowRecordBtn.Disable()
	}

	// initial PopUpChartWindowFlag
	PopUpChartWindowFlag := false

	// update show record btn
	ho.historyGUI.ShowRecordBtn.OnTapped = func() {

		if !PopUpChartWindowFlag {
			// set pop up chart window flag to true
			PopUpChartWindowFlag = true

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

			// []ntchart.chartPoint
			chartData := ntchart.ConvertFromDbToCheckpoint(dbTestEntries)

			// create testObject
			testObj, err := createTestObj(&sumData, chartData)
			if err != nil {
				errChan <- err
				return
			}

			// create holders for NewChartWindow (these holders are just required to call the "NewChartWindow" funciton. They won't be contrubiting in the function.)
			recordingFlag := true
			var p *ntPinger.Pinger

			// new chart window

			// if recordingFlag {
			// 	fmt.Println(testObj.GetChartData())

			// }

			NewChartWindow(a, testObj, &recordingFlag, p, db, entryChan, errChan, &PopUpChartWindowFlag)
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

				// delete select entry
				DelSelectedEntry(selectedEntries, selectedEntry{UUID: he.UUID, testType: he.TestType})

				// refresh table
				err = historyRefresh(a, w, db, entryChan, errChan, "ALL", selectedEntries, displayObjects)
				if err != nil {
					errChan <- err
				}
			}
		}, w)

		confirm.Show()
	}
}

// Func: history table refresh
func historyRefresh(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error, filter string, selectedEntries *[]selectedEntry, displayObjects *[]ntdb.HistoryEntry) error {

	// clean up all the items in ntGlobal.historyTable
	ntGlobal.historyTable.Objects = nil // Remove all child objects

	// reset displayObjects
	*displayObjects = []ntdb.HistoryEntry{}

	// get history entries
	historyEntries := GetHistoryEntries(db, errChan)

	// update history table body
	if len(*historyEntries) == 0 {
		return nil
	} else {
		for i := 0; i < len(*historyEntries); i++ {
			if filter != "ALL" {
				if (*historyEntries)[i].TestType != filter {
					continue
				}
			}
			// add display object
			*displayObjects = append(*displayObjects, (*historyEntries)[i])

			// add history table row
			go historyAddRow(a, w, &(*historyEntries)[i], ntGlobal.historyTable, db, entryChan, errChan, selectedEntries, displayObjects)
			// add some delays between each row to let the table sort by Id sequence
			time.Sleep(5 * time.Millisecond)
		}
	}
	return nil
}

// func: create testObject
func createTestObj(sumData *SummaryData, chartData *[]ntchart.ChartPoint) (testObject, error) {

	testType := sumData.Type

	switch testType {
	case "dns":
		obj := dnsObject{}
		obj.testSummary = sumData
		obj.ChartData = *chartData
		return &obj, nil
	case "http":
		obj := httpObject{}
		obj.testSummary = sumData
		obj.ChartData = *chartData
		return &obj, nil
	case "tcp":
		obj := tcpObject{}
		obj.testSummary = sumData
		obj.ChartData = *chartData
		return &obj, nil
	case "icmp":
		obj := icmpObject{}
		obj.testSummary = sumData
		obj.ChartData = *chartData
		return &obj, nil
	default:
		return nil, fmt.Errorf("testObject could not be created")
	}
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

// Func: select operation
func selectOperation(selectedEntries *[]selectedEntry, displayObjects *[]ntdb.HistoryEntry, selectAllCheckBoxFlag bool) {

	// update selectedEntries slice
	if len(*displayObjects) != 0 {
		if selectAllCheckBoxFlag {
			for _, s := range *displayObjects {
				AddSelectedEntry(selectedEntries, selectedEntry{UUID: s.UUID, testType: s.TestType})
			}
		} else {
			*selectedEntries = []selectedEntry{}
		}
	}
}

// Func: Get History Entries
func GetHistoryEntries(db *sql.DB, errChan chan error) *[]ntdb.HistoryEntry {
	// read the DB and obtain the historyEntries
	DbEntries, err := ntdb.ReadTableEntries(db, "history")
	if err != nil {
		errChan <- err
		return nil
	}

	// convert *[]DbEntry -> *[]ntdb.HistoryEntry
	historyEntries, err := ntdb.ConvertDbEntriesToHistoryEntries(DbEntries)
	if err != nil {
		errChan <- err
		return nil
	}

	return historyEntries
}
