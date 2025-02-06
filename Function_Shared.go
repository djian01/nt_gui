package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Function to create a vertical separator
func GUIVerticalSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.RGBA{200, 200, 200, 255}) // Gray color for a subtle look
	separator.SetMinSize(fyne.NewSize(1, 30))                        // 2px width, full height of row
	return separator
}
