package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"mvplayer/internal/layout"
)

const mediaDir = "/root/mvPlayer/.dsh-tmp/ff/media"

func testFiles() []string {
	return []string{
		filepath.Join(mediaDir, "h264_aac.mkv"),
		filepath.Join(mediaDir, "hevc_aac.mkv"),
		filepath.Join(mediaDir, "h264_ac3.mkv"),
		filepath.Join(mediaDir, "vp9_opus.mkv"),
	}
}

func newTestApp(t *testing.T) (*App, fyne.Window) {
	t.Helper()
	a := test.NewApp()
	a.Settings().SetTheme(NewTheme())
	w := test.NewWindow(nil)
	w.Resize(fyne.NewSize(1440, 900))
	app := New(a, w, filepath.Join(t.TempDir(), "history.json"))
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
	dir := os.Getenv("MVPLAYER_EVIDENCE")
	if dir == "" {
		dir = "/root/mvPlayer/.docs/evidence/tech-stack-research"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
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
	dir := os.Getenv("MVPLAYER_EVIDENCE")
	if dir == "" {
		dir = "/root/mvPlayer/.docs/evidence/tech-stack-research"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
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
