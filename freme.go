package main

import "fyne.io/fyne/v2"

type forcedVariant struct {
	fyne.Theme

	variant fyne.ThemeVariant
}

func DarkTheme(fallback fyne.Theme) fyne.Theme {
	return &forcedVariant{Theme: fallback, variant: 0} // avoid import loops
}

func LightTheme(fallback fyne.Theme) fyne.Theme {
	return &forcedVariant{Theme: fallback, variant: 1} // avoid import loops
}
