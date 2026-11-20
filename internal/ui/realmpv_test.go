//go:build cgo && mpv

package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"mvplayer/internal/player"
)

// TestRealMpvFramesIntoPanes is the end-to-end check: real libmpv decodes a
// 1080p file, the software renderer produces frames and the panes paint them.
func TestRealMpvFramesIntoPanes(t *testing.T) {
	os.Setenv("MVPLAYER_AO", "null")
	a := test.NewApp()
	a.Settings().SetTheme(NewTheme())
	w := test.NewWindow(nil)
	w.Resize(fyne.NewSize(1440, 900))
	app := New(a, w, filepath.Join(t.TempDir(), "history.json"))
	defer app.Stop()

	file := "/root/mvPlayer/.dsh-tmp/ff/media/hd1080.mkv"
	for i := 0; i < 2; i++ {
		app.AddFile(file)
	}
	// wait for libmpv to report a duration (file loaded)
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if app.panes[0].pl != nil && app.panes[0].pl.Duration() > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	app.stepOnce()
	for _, p := range app.panes {
		if p.Path == "" || p.pl == nil {
			continue
		}
		vr, ok := p.pl.(player.VideoRenderer)
		if !ok {
			t.Fatal("engine does not implement VideoRenderer")
		}
		sz := p.Root.Size()
		img, ok := vr.RenderFrame(int(sz.Width), int(sz.Height))
		if !ok {
			t.Fatalf("no frame for %s", p.Path)
		}
		p.SetFrame(img)
	}
	dir := "/root/mvPlayer/.docs/evidence/tech-stack-research"
	f, err := os.Create(filepath.Join(dir, "impl-realmpv.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
	if got := app.panes[0].pl.Duration(); got <= 0 {
		t.Fatalf("duration=%d, file did not load", got)
	}
}
