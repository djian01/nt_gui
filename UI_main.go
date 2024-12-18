package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// func makeUI: make the UI body
func makeUI(w fyne.Window, a fyne.App) {

	// set default frame
	fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())

	// ToolbarContainer
	tb := widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
	)

	ToolbarContainer := container.New(layout.NewVBoxLayout(), tb)

	// AppTabContainer
	AppTabContainer := container.NewAppTabs(
		container.NewTabItemWithIcon("Chart", theme.ComputerIcon(), Tag1Container(a, w)),
	)

	AppTabContainer.SetTabLocation(container.TabLocationLeading) // left

	// go routine: refresh the main container every 1 sec
	go func() {
		for {
			AppTabContainer.Refresh()
			time.Sleep(time.Second)
		}
	}()
	w.SetContent(mainContainer)

}
