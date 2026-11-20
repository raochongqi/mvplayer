package ui

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"mvplayer/internal/layout"
	"mvplayer/internal/player"
)

// testFiles returns four placeholder .mkv paths in a temp dir. Content is not
// needed: the unit tests use the in-memory engine, and the picker only filters
// by extension. Real playback is covered by the opt-in MVPLAYER_SAMPLE test.
var (
	mediaOnce sync.Once
	mediaDir  string
)

func mediaDirPath() string {
	mediaOnce.Do(func() {
		dir, err := os.MkdirTemp("", "mvplayer-media")
		if err != nil {
			panic(err)
		}
		mediaDir = dir
		for i := 1; i <= 4; i++ {
			_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.mkv", i)), []byte("mvplayer test"), 0o644)
		}
	})
	return mediaDir
}

func testFiles() []string {
	dir := mediaDirPath()
	out := make([]string, 0, 4)
	for i := 1; i <= 4; i++ {
		out = append(out, filepath.Join(dir, fmt.Sprintf("sample-%d.mkv", i)))
	}
	return out
}

func newTestApp(t *testing.T) (*App, fyne.Window) {
	t.Helper()
	a := test.NewApp()
	a.Settings().SetTheme(NewTheme())
	w := test.NewWindow(nil)
	w.Resize(fyne.NewSize(1440, 900))
	app := New(a, w, filepath.Join(t.TempDir(), "history.json"))
	// Unit tests use the deterministic fake engine; the real-mpv test swaps in
	// player.NewDefault explicitly.
	app.SetFactory(func() (player.Player, error) { return &fakeVideo{base: color.NRGBA{R: 40, G: 90, B: 160, A: 255}}, nil })
	t.Cleanup(app.Stop)
	return app, w
}

func TestAddFourFilesSwitchesToQuadLayout(t *testing.T) {
	app, _ := newTestApp(t)
	for i, f := range testFiles() {
		app.AddFile(f)
		if got := app.countPanes(); got != i+1 {
			t.Fatalf("after %d files countPanes=%d", i+1, got)
		}
	}
	if got := app.stage.Layout(); got != layout.Four {
		t.Fatalf("layout=%s want %s", got, layout.Four)
	}
	for _, p := range app.panes {
		if p.Path == "" {
			t.Fatal("expected all four panes filled")
		}
	}
}

func TestLayoutButtonsFollowPaneCount(t *testing.T) {
	app, _ := newTestApp(t)
	app.AddFile(testFiles()[0])
	if app.layoutBtns[layout.TwoV].Disabled() == false {
		// one pane: only the single layout must be enabled
		if !app.layoutBtns[layout.One].Disabled() {
			t.Fatal("single layout should stay enabled")
		}
	}
	app.AddFile(testFiles()[1])
	if !app.layoutBtns[layout.One].Disabled() {
		t.Fatal("1-up must be disabled with two panes")
	}
	if app.layoutBtns[layout.TwoV].Disabled() {
		t.Fatal("2-up must be enabled with two panes")
	}
}

func TestRemovePaneRelayouts(t *testing.T) {
	app, _ := newTestApp(t)
	for _, f := range testFiles() {
		app.AddFile(f)
	}
	app.removePane(app.panes[3])
	if got := app.countPanes(); got != 3 {
		t.Fatalf("countPanes=%d want 3", got)
	}
	if got := app.stage.Layout(); got != layout.ThreeL {
		t.Fatalf("layout=%s want %s", got, layout.ThreeL)
	}
}

func TestSplitterDragChangesRatio(t *testing.T) {
	app, _ := newTestApp(t)
	for _, f := range testFiles() {
		app.AddFile(f)
	}
	before := app.stage.Ratios()
	app.stage.drag(layout.SplitVertical, fyne.NewDelta(120, 0))
	after := app.stage.Ratios()
	if after.X <= before.X {
		t.Fatalf("ratio did not grow: %v -> %v", before, after)
	}
	if after.X > layout.MaxRatio || after.X < layout.MinRatio {
		t.Fatalf("ratio out of range: %v", after.X)
	}
}

func TestCaptureScreenshots(t *testing.T) {
	dir := evidenceDir(t)
	app, w := newTestApp(t)
	for _, f := range testFiles() {
		app.AddFile(f)
	}
	// let the ticker paint one frame of progress
	app.panes[0].SetProgress(754000, 6012000)
	save := func(name string) {
		img := w.Canvas().Capture()
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	}
	save("impl-4up.png")
	app.setLayout(layout.TwoV)
	save("impl-2v.png")
	app.setLayout(layout.ThreeQ)
	save("impl-3q.png")
}

func TestCaptureEmptyAndThreePaneStates(t *testing.T) {
	dir := evidenceDir(t)
	app, w := newTestApp(t)
	save := func(name string) {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, w.Canvas().Capture()); err != nil {
			t.Fatal(err)
		}
	}
	save("impl-empty.png")
	for _, f := range testFiles()[:3] {
		app.AddFile(f)
	}
	app.setLayout(layout.ThreeQ)
	save("impl-3q-empty-slot.png")
}

// evidenceDir returns a writable directory for screenshots. CI runs with an
// unprivileged user and no checkout paths, so the default is the test temp
// dir; set MVPLAYER_EVIDENCE to copy artefacts into the repo.
func evidenceDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("MVPLAYER_EVIDENCE")
	if dir == "" {
		dir = t.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}
