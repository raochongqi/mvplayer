package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

func TestShortcutSpacePausesAll(t *testing.T) {
	app, _ := newTestApp(t)
	for _, f := range testFiles()[:2] {
		app.AddFile(f)
	}
	app.onKey(&fyne.KeyEvent{Name: fyne.KeySpace})
	if !app.ctrl.Paused() {
		t.Fatal("space must pause all panes")
	}
	app.onKey(&fyne.KeyEvent{Name: fyne.KeySpace})
	if app.ctrl.Paused() {
		t.Fatal("second space must resume all panes")
	}
}

func TestShortcutArrowsSeekTenSeconds(t *testing.T) {
	app, _ := newTestApp(t)
	app.AddFile(testFiles()[0])
	pl := app.players[0]
	_ = pl.Seek(20_000)
	app.onKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if got := pl.Position(); got < 29_500 || got > 30_500 {
		t.Fatalf("right arrow pos=%d want ~30000", got)
	}
	app.onKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	if got := pl.Position(); got < 19_500 || got > 20_500 {
		t.Fatalf("left arrow pos=%d want ~20000", got)
	}
}

func TestMultiPickerListsVideos(t *testing.T) {
	app, _ := newTestApp(t)
	p := newMultiPicker(app.win, nil)
	p.navigate(storage.NewFileURI(mediaDir))
	n := 0
	for _, e := range p.entries {
		if !isDir(e) {
			n++
		}
	}
	if n < 4 {
		t.Fatalf("picker listed %d videos, want >=4", n)
	}
	for _, e := range p.entries {
		if isDir(e) && e.Name() == "" {
			t.Fatal("empty dir name")
		}
	}
}
