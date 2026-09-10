package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Icon.png
var iconPNG []byte

var resourceIconPng = fyne.NewStaticResource("Icon.png", iconPNG)
