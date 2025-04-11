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

func HTTPPingContainer(a fyne.App, w fyne.Window, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) *fyne.Container {

	// index
	ntGlobal.httpIndex = 1

	// ** Action-Button Card **
	httpPingAddBtn := widget.NewButtonWithIcon("Add HTTP Ping", theme.ContentAddIcon(), func() {})
	httpPingAddBtn.Importance = widget.HighImportance
	httpPingAddBtnContainer := container.New(layout.NewBorderLayout(nil, nil, httpPingAddBtn, nil), httpPingAddBtn)
	httpPingAddBtncard := widget.NewCard("", "", httpPingAddBtnContainer)

	// ** Table Container **
	httpHeader := httpGUIRow{}
	httpHeader.Initial()
	httpHeaderRow := httpHeader.GenerateHeaderRow()

	httpTableBody := container.New(layout.NewVBoxLayout())
	ntGlobal.httpTable = httpTableBody

	httpTableScroll := container.NewScroll(httpTableBody)
	httpTableContainer := container.New(layout.NewBorderLayout(httpHeaderRow, nil, nil, nil), httpHeaderRow, httpTableScroll)

	// ** Table Card **
	httpTableCard := widget.NewCard("", "", httpTableContainer)

	// ** Main Container **
	HTTPSpaceHolder := widget.NewLabel("    ")
	HTTPMainContainerInner := container.New(layout.NewBorderLayout(httpPingAddBtncard, nil, nil, nil), httpPingAddBtncard, httpTableCard)
	HTTPMainContainerOuter := container.New(layout.NewBorderLayout(HTTPSpaceHolder, HTTPSpaceHolder, HTTPSpaceHolder, HTTPSpaceHolder), HTTPSpaceHolder, HTTPMainContainerInner)

	// dnsPingAddBtn action
	httpPingAddBtn.OnTapped = func() {
		// Initial a New Test
		NewTest(a, "http", db, entryChan, errChan)
	}

	// Return your DNS ping interface components here
	return HTTPMainContainerOuter // Temporary empty container, replace with your actual UI

}
