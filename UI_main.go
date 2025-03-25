package main

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

// Global Var: nt table
var ntGlobal ntGUIGlboal

// func makeUI: make the UI body
func makeUI(w fyne.Window, a fyne.App, db *sql.DB, entryChan chan ntdb.DbEntry) {

	// set theme variable
	currentTheme := "light"
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantLight {
		currentTheme = "light"
	} else {
		currentTheme = "dark"
	}

	// ToolbarContainer
	ToolbarWidget := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.RadioButtonCheckedIcon(), func() {
			// Toggle between light and dark theme
			if currentTheme == "light" {
				currentTheme = "dark"
				fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			} else {
				currentTheme = "light"
				fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			}
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			// Create about dialog with left-aligned text
			aboutContent := container.NewVBox(
				widget.NewLabel("Version:  1.0.0"),
				widget.NewLabel("Developed By:   Dennis Jian"),
				container.NewHBox(
					widget.NewLabel("Project Home: "),
					widget.NewHyperlink("https://github.com/djian01/nt_gui",
						parseURL("https://github.com/djian01/nt_gui")),
				),
				widget.NewLabel(""), // Add a blank line
			)

			aboutOkButton := widget.NewButton("           OK          ", nil)
			aboutOkButton.Importance = widget.HighImportance

			aboutDialog := dialog.NewCustom(
				"About NT (Net-Test) GUI",
				"", // Empty string since we'll use our custom button
				container.NewVBox(
					aboutContent,
					container.NewPadded(container.NewCenter(aboutOkButton)),
				),
				w)

			aboutOkButton.OnTapped = func() {
				aboutDialog.Hide()
			}

			aboutDialog.Resize(fyne.NewSize(500, 100))
			aboutDialog.Show()
		}),
	)

	ToolbarContainer := container.New(layout.NewVBoxLayout(), ToolbarWidget)

	// initial history entries
	historyEntries := []ntdb.HistoryEntry{}

	// Create resource from SVG file
	icmpIcon := theme.NewThemedResource(resourceIcmpIconSvg)
	tcpIcon := theme.NewThemedResource(resourceTcpIconSvg)
	dnsIcon := theme.NewThemedResource(resourceDnsIconSvg)
	httpIcon := theme.NewThemedResource(resourceHttpIconSvg)
	analyIcon := theme.NewThemedResource(resourceAnalyIconSvg)
	historyIcon := theme.NewThemedResource(resourceHistoryIconSvg)

	// AppTabContainer
	AppTabContainer := container.NewAppTabs(
		container.NewTabItemWithIcon("ICMP Ping", icmpIcon, ICMPPingContainer(a, w, db, entryChan)),
		container.NewTabItemWithIcon("TCP Ping", tcpIcon, TCPPingContainer(a, w, db, entryChan)),
		container.NewTabItemWithIcon("HTTP Ping", httpIcon, HTTPPingContainer(a, w, db, entryChan)),
		container.NewTabItemWithIcon("DNS Ping", dnsIcon, DNSPingContainer(a, w, db, entryChan)),
		container.NewTabItemWithIcon("Result Analysis", analyIcon, ResultAnalysisContainer(a, w, db, entryChan)),
		container.NewTabItemWithIcon("History", historyIcon, HistoryContainer(a, w, &historyEntries, db, entryChan)),
	)

	AppTabContainer.SetTabLocation(container.TabLocationLeading) // left

	AppTabContainer.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "History" {
			// refresh History table
			err := historyRefresh(a, w, &historyEntries, db, entryChan)
			if err != nil {
				logger.Println(err)
			}
		}
	}

	// MainContainer
	mainContainer := container.NewBorder(ToolbarContainer, nil, nil, nil, AppTabContainer)

	w.SetContent(mainContainer)

}
