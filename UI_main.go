package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
		widget.NewToolbarAction(theme.ColorPaletteIcon(), func() {
			// Toggle between light and dark theme
			if currentTheme == "light" {
				fmt.Println("set dark theme")
				currentTheme = "dark"
				fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			} else {
				fmt.Println("set light theme")
				currentTheme = "light"
				fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			}
		}),
		// widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
	)

	ToolbarContainer := container.New(layout.NewVBoxLayout(), ToolbarWidget)

	// AppTabContainer
	AppTabContainer := container.NewAppTabs(
		container.NewTabItemWithIcon("ICMP Ping", theme.ComputerIcon(), ICMPPingContainer(a, w)),
		container.NewTabItemWithIcon("TCP Ping", theme.ComputerIcon(), TCPPingContainer(a, w)),
		container.NewTabItemWithIcon("HTTP Ping", theme.ComputerIcon(), HTTPPingContainer(a, w)),
		container.NewTabItemWithIcon("DNS Ping", theme.ComputerIcon(), DNSPingContainer(a, w)),
		container.NewTabItemWithIcon("Result Analysis", theme.ComputerIcon(), ResultAnalysisContainer(a, w)),
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
