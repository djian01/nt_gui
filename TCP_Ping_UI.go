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

func TCPPingContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) *fyne.Container {

	// index
	ntGlobal.tcpIndex = 1

	// ** Add-Button Card **
	tcpPingAddBtn := widget.NewButtonWithIcon("Add TCP Ping", theme.ContentAddIcon(), func() {})
	tcpPingAddBtn.Importance = widget.HighImportance
	tcpPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, tcpPingAddBtn, nil), tcpPingAddBtn)
	tcpPingAddBtncard := widget.NewCard("", "", tcpPingAddBtnContainer)

	// ** Table Container **
	tcpHeader := tcpGUIRow{}
	tcpHeader.Initial()
	tcpHeaderRow := tcpHeader.GenerateHeaderRow()

	tcpTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.tcpTable = tcpTableBody

	tcpTableScroll := container.NewScroll(tcpTableBody)
	tcpTableContainer := container.New(layout.NewBorderLayout(tcpHeaderRow, nil, nil, nil), tcpHeaderRow, tcpTableScroll)

	// ** Table Card **
	tcpTableCard := widget.NewCard("", "", tcpTableContainer)

	// ** Main Container **
	tcpSpaceHolder := widget.NewLabel("    ")
	tcpMainContainerInner := container.New(layout.NewBorderLayout(tcpPingAddBtncard, nil, nil, nil), tcpPingAddBtncard, tcpTableCard)
	tcpMainContainerOuter := container.New(layout.NewBorderLayout(tcpSpaceHolder, tcpSpaceHolder, tcpSpaceHolder, tcpSpaceHolder), tcpSpaceHolder, tcpMainContainerInner)

	// dnsPingAddBtn action
	tcpPingAddBtn.OnTapped = func() {
		// Initial a New Test
		NewTest(a, "tcp", db, entryChan, errChan)
	}

	// Return your DNS ping interface components here
	return tcpMainContainerOuter // Temporary empty container, replace with your actual UI

}
