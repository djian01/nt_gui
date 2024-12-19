package main

import (
	"net/url"
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

func parseURL(urlStr string) *url.URL {
	link, _ := url.Parse(urlStr)
	return link
}
