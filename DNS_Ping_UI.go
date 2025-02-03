package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func DNSPingContainer(a fyne.App, w fyne.Window) *fyne.Container {

	// Add Button Card
	DNSPingAddBtn := widget.NewButtonWithIcon("Add DNS Ping", theme.ContentAddIcon(), func() {})
	DNSPingAddBtn.Importance = widget.HighImportance
	DNSPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, DNSPingAddBtn, nil), DNSPingAddBtn)
	DNSPingAddBtncard := widget.NewCard("", "", DNSPingAddBtnContainer)

	// table
	header := []string{"header 1", "header 2", "header 3", "header 4", "header 5", "header 6"}
	data := [][]string{
		{"a", "b", "c", "d", "e", "f"},
		{"g", "h", "i", "j0000000000000000000", "k", "l"},
		{"m", "n", "o", "p", "q", "r"},
		{"s", "t", "u", "v", "w", "x"},
		{"a", "b", "c", "d", "e", "f"},
		{"g", "h", "i", "j0000000000000000000", "k", "l"},
		{"m", "n", "o", "p", "q", "r"},
		{"s", "t", "u", "v", "w", "x"},
		{"a", "b", "c", "d", "e", "f"},
		{"g", "h", "i", "j0000000000000000000", "k", "l"},
		{"m", "n", "o", "p", "q", "r"},
		{"s", "t", "u", "v", "w", "x"},
		{"a", "b", "c", "d", "e", "f"},
		{"g", "h", "i", "j0000000000000000000", "k", "l"},
		{"m", "n", "o", "p", "q", "r"},
		{"s", "t", "u", "v", "w", "x"},
		{"a", "b", "c", "d", "e", "f"},
		{"g", "h", "i", "j0000000000000000000", "k", "l"},
		{"m", "n", "o", "p", "q", "r"},
		{"s", "t", "u", "v", "w", "x"},
	}
	tableBody := [][]string{}
	tableBody = append(tableBody, header)
	tableBody = append(tableBody, data...)

	// widget.NewTable(Length(){},CreateCell(){},UpdateCell(){})
	t := widget.NewTable(

		// callback - length
		func() (int, int) {
			return len(tableBody), len(data[0])
		},

		// callback - CreateCell
		func() fyne.CanvasObject {
			// the basic Cell Width is defined by the string length below
			item := widget.NewLabel("AAAAAAAAA")
			item.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			return item
		},

		// callback - UpdateCell
		//// cell: contains Row, Col int for a cell, item: the CanvasObject from CreateCell Callback
		func(id widget.TableCellID, item fyne.CanvasObject) {
			if id.Row == 0 {
				// Header row
				item.(*widget.Label).TextStyle.Bold = true
				item.(*widget.Label).SetText(tableBody[id.Row][id.Col])
			} else {
				item.(*widget.Label).SetText(fmt.Sprintf("Cell[%v][%v]: %s ", id.Row, id.Col, tableBody[id.Row][id.Col]))
			}

		},
	)

	t.StickyRowCount = 1

	t.Select(widget.TableCellID{
		Row: 1,
		Col: 2,
	})

	t.OnSelected = func(id widget.TableCellID) {
		fmt.Printf("Selected: [%v][%v], %s \n", id.Row, id.Col, tableBody[id.Row][id.Col])
	}

	// set width for all colunms width for loop
	for i := 0; i < len(data[0]); i++ {
		t.SetColumnWidth(i, 150)
	}

	// set the width for one column
	t.SetColumnWidth(0, 120)

	// button to change data[0][0], update table display
	// bt := widget.NewButton("Change data[0][0]", func() {
	// 	data[0][0] = "ABC"
	// 	t.Refresh()
	// })

	tableContainer := container.NewHScroll(t)

	tableCard := widget.NewCard("", "", tableContainer)

	//// Main Container
	DNSSpaceHolder := widget.NewLabel("    ")
	DNSMainContainerInner := container.New(layout.NewBorderLayout(DNSPingAddBtncard, nil, nil, nil), DNSPingAddBtncard, tableCard)
	DNSMainContainerOuter := container.New(layout.NewBorderLayout(DNSSpaceHolder, DNSSpaceHolder, DNSSpaceHolder, DNSSpaceHolder), DNSSpaceHolder, DNSMainContainerInner)

	// Return your DNS ping interface components here
	return DNSMainContainerOuter // Temporary empty container, replace with your actual UI

}
