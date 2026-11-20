package main

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"mvplayer/internal/ui"
)

func main() {
	a := app.NewWithID("io.github.mvplayer")
	a.Settings().SetTheme(ui.NewTheme())
	a.SetIcon(ui.Icon)

	w := a.NewWindow("mvPlayer")
	w.Resize(fyne.NewSize(1280, 800))
	w.CenterOnScreen()

	cfg, err := os.UserConfigDir()
	if err != nil {
		cfg = "."
	}
	histPath := filepath.Join(cfg, "mvPlayer", "history.json")

	ui.New(a, w, histPath).Start()
}
