package ntwidget

import "image/color"

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func isDarkBackground(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	luma := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 257.0
	return luma < 120
}
