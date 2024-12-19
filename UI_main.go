package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// func makeUI: make the UI body
func makeUI(w fyne.Window, a fyne.App) {

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
			aboutContent := widget.NewLabel(
				"Version: 1.0.0\n" +
					"Developed By: Dennis Jian\n" +
					"Project Home: https://github.com/djian01/nt_gui\n\n")
			aboutContent.Alignment = fyne.TextAlignLeading // Left alignment

			aboutDialog := dialog.NewCustom(
				"About NT (Net-Test) GUI",
				"OK",
				aboutContent,
				w)
			aboutDialog.Resize(fyne.NewSize(500, 130))
			aboutDialog.Show()
		}),
	)

	ToolbarContainer := container.New(layout.NewVBoxLayout(), ToolbarWidget)

	// Create resource from SVG file
	icmpIcon := theme.NewThemedResource(resourceIcmpIconSvg)
	tcpIcon := theme.NewThemedResource(resourceTcpIconSvg)
	dnsIcon := theme.NewThemedResource(resourceDnsIconSvg)
	httpIcon := theme.NewThemedResource(resourceHttpIconSvg)
	analyIcon := theme.NewThemedResource(resourceAnalyIconSvg)

	// AppTabContainer
	AppTabContainer := container.NewAppTabs(
		container.NewTabItemWithIcon("ICMP Ping", icmpIcon, ICMPPingContainer(a, w)),
		container.NewTabItemWithIcon("TCP Ping", tcpIcon, TCPPingContainer(a, w)),
		container.NewTabItemWithIcon("HTTP Ping", httpIcon, HTTPPingContainer(a, w)),
		container.NewTabItemWithIcon("DNS Ping", dnsIcon, DNSPingContainer(a, w)),
		container.NewTabItemWithIcon("Result Analysis", analyIcon, ResultAnalysisContainer(a, w)),
	)

	AppTabContainer.SetTabLocation(container.TabLocationLeading) // left

	// MainContainer
	mainContainer := container.NewBorder(ToolbarContainer, nil, nil, nil, AppTabContainer)

	// go routine: refresh the main container every 1 sec
	go func() {
		for {
			AppTabContainer.Refresh()
			time.Sleep(time.Second)
		}
	}()

	w.SetContent(mainContainer)

}
