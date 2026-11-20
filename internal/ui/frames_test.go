package ui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"mvplayer/internal/history"
	"mvplayer/internal/player"
)

// fakeVideo paints a coloured gradient so a screenshot proves frames reach
// the canvas without needing libmpv in unit tests.
type fakeVideo struct {
	pos, dur int64
	paused   bool
	rate     float64
	base     color.NRGBA
}

func (f *fakeVideo) Open(string) error          { return nil }
func (f *fakeVideo) Close() error               { return nil }
func (f *fakeVideo) Play() error                { f.paused = false; return nil }
func (f *fakeVideo) Pause() error               { f.paused = true; return nil }
func (f *fakeVideo) Paused() bool               { return f.paused }
func (f *fakeVideo) SeekTo(ms int64) error        { f.pos = ms; return nil }
func (f *fakeVideo) Position() int64            { return f.pos }
func (f *fakeVideo) Duration() int64            { return 600000 }
func (f *fakeVideo) SetRate(r float64) error    { f.rate = r; return nil }
func (f *fakeVideo) Rate() float64              { return f.rate }
func (f *fakeVideo) RenderFrame(w, h int) (image.Image, bool) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			k := uint8((x + y) * 255 / (w + h))
			img.SetRGBA(x, y, color.RGBA{R: f.base.R, G: k, B: f.base.B, A: 255})
		}
	}
	return img, true
}

func TestFramesPaintIntoPanes(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(NewTheme())
	w := test.NewWindow(nil)
	w.Resize(fyne.NewSize(1440, 900))

	app := New(a, w, filepath.Join(t.TempDir(), "history.json"))
	defer app.Stop()
	colors := []color.NRGBA{{200, 0, 0, 255}, {0, 60, 200, 255}, {0, 160, 90, 255}, {210, 150, 0, 255}}
	i := 0
	app.SetFactory(func() (player.Player, error) {
		f := &fakeVideo{base: colors[i%len(colors)]}
		i++
		return f, nil
	})
	for _, f := range testFiles() {
		app.AddFile(f)
	}
	// drive one render pass synchronously
	for idx, p := range app.panes {
		if p.Path == "" || idx >= len(app.players) {
			continue
		}
		vr := app.players[idx].(player.VideoRenderer)
		sz := p.Root.Size()
		if img, ok := vr.RenderFrame(int(sz.Width), int(sz.Height)); ok {
			p.SetFrame(img)
		}
	}
	dir := evidenceDir(t)
	f, err := os.Create(filepath.Join(dir, "impl-frames.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}

func historyEntry(path string, size, mtime, pos int64) history.Entry {
	return history.Entry{Path: path, Size: size, ModTime: mtime, PositionMs: pos, DurationMs: 600000, UpdatedAt: 1}
}

func TestHistoryRestoresPosition(t *testing.T) {
	app, _ := newTestApp(t)
	file := testFiles()[0]
	st, err := fileStat(file)
	if err != nil {
		t.Fatal(err)
	}
	app.hist.Put(historyEntry(file, st.size, st.mtime, 120_000))
	app.AddFile(file)
	app.stepOnce()
	if got := app.players[0].Position(); got < 119_000 || got > 121_000 {
		t.Fatalf("restored position=%d want ~120000", got)
	}
}

func TestRemoveMiddlePaneKeepsOthersWired(t *testing.T) {
	app, _ := newTestApp(t)
	for _, f := range testFiles() {
		app.AddFile(f)
	}
	p2, p3 := app.panes[2], app.panes[3]
	_ = p2.pl.SeekTo(5_000)
	_ = p3.pl.SeekTo(7_000)
	app.removePane(app.panes[1])
	if p2.pl == nil || p3.pl == nil {
		t.Fatal("removal detached another pane")
	}
	app.nudgePane(p2, 10_000)
	if got := p2.pl.Position(); got != 15_000 {
		t.Fatalf("pane2 pos=%d want 15000", got)
	}
	if got := p3.pl.Position(); got != 7_000 {
		t.Fatalf("pane3 pos=%d want 7000 (must not move)", got)
	}
	if app.ctrl.Len() != 3 {
		t.Fatalf("controller len=%d want 3", app.ctrl.Len())
	}
	// the empty slot can take a new file and bind to that pane only
	app.AddFile(testFiles()[1])
	if app.panes[1].pl == nil || app.panes[1].pl == p2.pl {
		t.Fatal("new file not bound to the emptied pane")
	}
}
