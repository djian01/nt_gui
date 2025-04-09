package main

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

func ICMPPingContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) *fyne.Container {

	// index
	ntGlobal.icmpIndex = 1

	// ** Add-Button Card **
	icmpPingAddBtn := widget.NewButtonWithIcon("Add ICMP Ping", theme.ContentAddIcon(), func() {})
	icmpPingAddBtn.Importance = widget.HighImportance
	icmpPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, icmpPingAddBtn, nil), icmpPingAddBtn)
	icmpPingAddBtncard := widget.NewCard("", "", icmpPingAddBtnContainer)

	// ** Table Container **
	icmpHeader := icmpGUIRow{}
	icmpHeader.Initial()
	icmpHeaderRow := icmpHeader.GenerateHeaderRow()

	icmpTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.icmpTable = icmpTableBody

	icmpTableScroll := container.NewScroll(icmpTableBody)
	icmpTableContainer := container.New(layout.NewBorderLayout(icmpHeaderRow, nil, nil, nil), icmpHeaderRow, icmpTableScroll)

	// ** Table Card **
	icmpTableCard := widget.NewCard("", "", icmpTableContainer)

	// ** Main Container **
	icmpSpaceHolder := widget.NewLabel("    ")
	icmpMainContainerInner := container.New(layout.NewBorderLayout(icmpPingAddBtncard, nil, nil, nil), icmpPingAddBtncard, icmpTableCard)
	icmpMainContainerOuter := container.New(layout.NewBorderLayout(icmpSpaceHolder, icmpSpaceHolder, icmpSpaceHolder, icmpSpaceHolder), icmpSpaceHolder, icmpMainContainerInner)

	// dnsPingAddBtn action
	icmpPingAddBtn.OnTapped = func() {
		// Initial a New Test
		NewTest(a, "icmp", db, entryChan, errChan)
	}

	// Return your DNS ping interface components here
	return icmpMainContainerOuter // Temporary empty container, replace with your actual UI

}
