package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconBytes []byte

// Icon is the window/taskbar icon.
var Icon = fyne.NewStaticResource("icon.png", iconBytes)
