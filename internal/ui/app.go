package ui

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2/canvas"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"mvplayer/internal/history"
	"mvplayer/internal/layout"
	"mvplayer/internal/player"
)

// App wires the window, controls and players together.
type App struct {
	fyneApp fyne.App
	win     fyne.Window
	stage   *Stage
	ctrl    *player.Controller
	panes   []*Pane
	players []player.Player

	hist     *history.Store
	histPath string

	overlay     float64
	overlayStep float64
	globalStep  int
	layoutBtns map[layout.Layout]*chipButton
	layoutBox  *fyne.Container
	pauseBtn   *chipButton
	speedSel   *widget.Select
	stepSel    *widget.Select
	rateLbl    *widget.Label
	root       *fyne.Container
	statusLbl  *widget.Label
	stop       chan struct{}
	stopOnce   sync.Once
}

// New builds the application UI on top of the given window.
func New(a fyne.App, w fyne.Window, histPath string) *App {
	app := &App{
		fyneApp:    a,
		win:        w,
		stage:      NewStage(),
		ctrl:       player.NewController(),
		hist:       history.New(),
		histPath:   histPath,
		overlay:     0,
		overlayStep: 0.25,
		globalStep:  10,
		layoutBtns: map[layout.Layout]*chipButton{},
		stop:       make(chan struct{}),
	}
	if s, err := history.Load(histPath); err == nil {
		app.hist = s
	}
	for i := 0; i < 4; i++ {
		pane := NewPane(i, PaneOptions{
			OnRemove: app.removePane,
			OnRate:   app.setPaneRate,
			OnStep:   func(p *Pane, n int) {},
			OnToggle: app.togglePane,
			OnNudge:  app.nudgePane,
		})
		pane.SetSeekHandler(app.seekPaneAbsolute)
		app.panes = append(app.panes, pane)
	}
	app.stage.SetPanes(app.panes)
	app.stage.SetLayout(layout.One)

	app.root = container.NewBorder(app.buildToolbar(), app.buildFooter(), nil, nil, app.stage)
	w.SetContent(app.root)
	w.SetOnDropped(app.onDropped)
	w.SetCloseIntercept(func() {
		app.persistHistory()
		w.Close()
	})
	go app.tick()
	return app
}

func (a *App) buildToolbar() fyne.CanvasObject {
	openBtn := newChip("＋ 打开视频", a.openDialog)
	openBtn.SetPrimary(true)
	histBtn := newChip("≡ 历史", a.showHistory)

	a.layoutBox = container.NewHBox()
	labels := map[layout.Layout]string{
		layout.One: "1", layout.TwoV: "∥", layout.TwoH: "＝",
		layout.ThreeL: "◧", layout.ThreeR: "◨", layout.ThreeQ: "▣", layout.Four: "▦",
	}
	for _, l := range layout.All {
		l := l
		b := newChip(labels[l], func() { a.setLayout(l) })
		a.layoutBtns[l] = b
		a.layoutBox.Add(container.NewGridWrap(fyne.NewSize(30, 26), b))
	}

	prevBtn := newChip("«", func() { a.seekAll(-int64(a.globalStep * 1000)) })
	nextBtn := newChip("»", func() { a.seekAll(int64(a.globalStep * 1000)) })
	a.stepSel = widget.NewSelect([]string{"10s", "30s", "60s", "120s"}, func(s string) {
		if n, ok := parseStep(s); ok {
			a.globalStep = n
		}
	})
	a.stepSel.SetSelected("10s")
	a.pauseBtn = newChip("Ⅱ 全部暂停", a.toggleAll)

	a.rateLbl = widget.NewLabel("基线 1.00×　叠加 +0.00")
	a.speedSel = widget.NewSelect(presetRates(), nil)
	a.speedSel.SetSelected("1.0×")
	a.speedSel.OnChanged = func(s string) {
		if r, ok := parseRate(s); ok {
			a.setAllRates(r)
		}
	}
	minusBtn := newChip("−", func() { a.nudgeAllRates(-a.overlayStep) })
	plusBtn := newChip("＋", func() { a.nudgeAllRates(a.overlayStep) })

	settingsBtn := newChip("⚙", func() {
		dialog.ShowInformation("设置", "自定义倍速/档位与历史开关将在下一版接入。", a.win)
	})

	vsep := func() fyne.CanvasObject { return container.NewGridWrap(fyne.NewSize(1, 22), canvas.NewRectangle(ColorLine)) }
	left := container.NewHBox(openBtn, histBtn, vsep(), a.layoutBox)
	center := container.NewHBox(widget.NewLabel("全局"), prevBtn,
		container.NewGridWrap(fyne.NewSize(58, 26), a.stepSel), nextBtn, a.pauseBtn)
	overlayStepSel := widget.NewSelect([]string{"0.1", "0.25", "0.5", "1.0"}, func(s string) {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			a.overlayStep = v
		}
	})
	overlayStepSel.SetSelected("0.25")
	right := container.NewHBox(widget.NewLabel("全局倍速"),
		container.NewGridWrap(fyne.NewSize(66, 26), a.speedSel),
		widget.NewLabel("叠加"), minusBtn,
		container.NewGridWrap(fyne.NewSize(54, 26), overlayStepSel), plusBtn, a.rateLbl, settingsBtn)
	bar := container.NewBorder(nil, nil, left, right, container.NewCenter(center))
	return container.NewStack(canvas.NewRectangle(ColorChrome), container.NewPadded(bar))
}

func (a *App) buildFooter() fyne.CanvasObject {
	a.statusLbl = widget.NewLabel("0 路 · 叠加 +0.00×")
	hint := widget.NewLabel("空格 全局暂停　←/→ 全局 ±10s　拖动中缝调整布局")
	row := container.NewBorder(nil, nil, a.statusLbl, hint)
	return container.NewStack(canvas.NewRectangle(ColorChrome), container.NewPadded(row))
}

// Start restores the window size and shows it.
func (a *App) Start() {
	a.refreshLayoutButtons()
	a.win.ShowAndRun()
}

func (a *App) openDialog() {
	d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		defer rc.Close()
		a.AddFile(rc.URI().Path())
	}, a.win)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".mkv", ".mp4", ".webm"}))
	d.Show()
}

// AddFile adds one video to the first empty pane.
func (a *App) AddFile(path string) {
	idx := -1
	for i, p := range a.panes {
		if p.Path == "" {
			idx = i
			break
		}
	}
	if idx < 0 {
		dialog.ShowInformation("已是 4 路", "请先移除一路再打开新文件。", a.win)
		return
	}
	eng, err := player.NewDefault()
	if err != nil {
		dialog.ShowError(fmt.Errorf("创建播放器失败：%w", err), a.win)
		return
	}
	if err := eng.Open(path); err != nil {
		dialog.ShowError(fmt.Errorf("打开失败：%w", err), a.win)
		return
	}
	a.players = append(a.players, eng)
	a.ctrl.Add(eng)
	pane := a.panes[idx]
	pane.SetFile(path)
	pane.SetRateBadge(1.0)
	pane.SetPlaying(true)
	_ = eng.Play()
	a.stage.SetPanes(a.panes[:a.countPanes()])
	a.relayoutForCount()
	a.refreshLayoutButtons()
}

func (a *App) countPanes() int {
	n := 0
	for _, p := range a.panes {
		if p.Path != "" {
			n++
		}
	}
	if n < 1 {
		n = 1
	}
	return n
}

func (a *App) relayoutForCount() {
	n := a.countPanes()
	if !layout.ValidFor(a.stage.Layout(), n) {
		a.stage.SetLayout(layout.DefaultFor(n))
		a.refreshLayoutButtons()
	}
}

func (a *App) setLayout(l layout.Layout) {
	if !layout.ValidFor(l, a.countPanes()) {
		return
	}
	a.stage.SetLayout(l)
	a.refreshLayoutButtons()
}

func (a *App) refreshLayoutButtons() {
	n := a.countPanes()
	cur := a.stage.Layout()
	for l, b := range a.layoutBtns {
		b.SetEnabled(layout.ValidFor(l, n))
		b.SetActive(l == cur)
	}
}

func (a *App) removePane(p *Pane) {
	i := p.Index
	if i < 0 || i >= len(a.panes) || a.panes[i] != p {
		return
	}
	// remove the player at the same logical slot
	if i < len(a.players) {
		_ = a.players[i].Close()
		a.ctrl.RemoveAt(i)
		a.players = append(a.players[:i], a.players[i+1:]...)
	}
	p.Clear()
	a.stage.SetPanes(a.panes[:a.countPanes()])
	a.relayoutForCount()
	a.refreshLayoutButtons()
}

func (a *App) setPaneRate(p *Pane, rate float64) {
	if i := a.indexOf(p); i >= 0 && i < len(a.players) {
		_ = a.players[i].SetRate(player.ClampRate(rate))
		p.SetRateBadge(player.ClampRate(rate))
	}
}

func (a *App) togglePane(p *Pane) {
	if i := a.indexOf(p); i >= 0 && i < len(a.players) {
		pl := a.players[i]
		if pl.Paused() {
			_ = pl.Play()
		} else {
			_ = pl.Pause()
		}
	}
}

func (a *App) nudgePane(p *Pane, ms int) {
	if i := a.indexOf(p); i >= 0 && i < len(a.players) {
		pl := a.players[i]
		target := pl.Position() + int64(ms)
		if target < 0 {
			target = 0
		}
		if d := pl.Duration(); d > 0 && target > d {
			target = d
		}
		_ = pl.Seek(target)
	}
}

func (a *App) seekPaneAbsolute(p *Pane, f float32) {
	if i := a.indexOf(p); i >= 0 && i < len(a.players) {
		pl := a.players[i]
		if d := pl.Duration(); d > 0 {
			_ = pl.Seek(int64(float32(d) * f))
		}
	}
}

func (a *App) indexOf(p *Pane) int {
	for i, q := range a.panes {
		if q == p {
			return i
		}
	}
	return -1
}

func (a *App) toggleAll() {
	paused, _ := a.ctrl.ToggleAll()
	if paused {
		a.pauseBtn.SetText("▶ 全部播放")
	} else {
		a.pauseBtn.SetText("Ⅱ 全部暂停")
	}
}

func (a *App) seekAll(ms int64) { _ = a.ctrl.SeekAll(ms) }

func (a *App) setAllRates(r float64) {
	_ = a.ctrl.SetBaseAll(r)
	for _, p := range a.panes {
		if p.Path != "" {
			p.SetRateBadge(r + a.overlay)
		}
	}
	a.refreshRateLabel()
}

func (a *App) nudgeAllRates(delta float64) {
	a.overlay += delta
	a.overlay = clampOverlay(a.overlay)
	for i, p := range a.panes {
		if p.Path != "" && i < len(a.players) {
			r := player.ClampRate(a.players[i].Rate() + delta)
			_ = a.players[i].SetRate(r)
			p.SetRateBadge(r)
		}
	}
	a.refreshRateLabel()
}

func clampOverlay(v float64) float64 {
	if v > 7.5 {
		return 7.5
	}
	if v < -0.5 {
		return -0.5
	}
	return v
}

func (a *App) refreshRateLabel() {
	if a.rateLbl != nil {
		a.rateLbl.SetText(fmt.Sprintf("叠加 %+.2f×", a.overlay))
	}
}

func (a *App) persistHistory() {
	for i, p := range a.panes {
		if p.Path == "" || i >= len(a.players) {
			continue
		}
		pl := a.players[i]
		st, err := fileStat(p.Path)
		if err != nil {
			continue
		}
		a.hist.Put(history.Entry{
			Path:       p.Path,
			Size:       st.size,
			ModTime:    st.mtime,
			PositionMs: pl.Position(),
			DurationMs: pl.Duration(),
			UpdatedAt:  history.Now(),
		})
	}
	a.hist.Prune(200)
	_ = a.hist.Save(a.histPath)
}

func (a *App) showHistory() {
	if len(a.hist.Entries) == 0 {
		dialog.ShowInformation("历史", "还没有播放记录。", a.win)
		return
	}
	var items []string
	for _, e := range a.hist.Entries {
		items = append(items, fmt.Sprintf("%s　%s", shortName(e.Path), clock(e.PositionMs)))
	}
	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(items[i]) },
	)
	dialog.ShowCustom("历史播放", "关闭", list, a.win)
}

func (a *App) onDropped(_ fyne.Position, uris []fyne.URI) {
	for _, u := range uris {
		a.AddFile(u.Path())
	}
}

func (a *App) refreshStatus() {
	if a.statusLbl == nil {
		return
	}
	n := 0
	for _, p := range a.panes {
		if p.Path != "" {
			n++
		}
	}
	state := "播放中"
	if a.ctrl.Paused() && n > 0 {
		state = "已暂停"
	}
	a.statusLbl.SetText(fmt.Sprintf("%d 路 · %s · 叠加 %+.2f×", n, state, a.overlay))
}

func (a *App) Stop() {
	a.stopOnce.Do(func() { close(a.stop) })
}

func (a *App) tick() {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-t.C:
		}
		a.refreshStatus()
		for i, p := range a.panes {
			if p.Path == "" || i >= len(a.players) {
				continue
			}
			pl := a.players[i]
			p.SetProgress(pl.Position(), pl.Duration())
			p.SetPlaying(!pl.Paused())
		}
	}
}

type stat struct {
	size  int64
	mtime int64
}

func fileStat(path string) (stat, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return stat{}, err
	}
	return stat{size: fi.Size(), mtime: fi.ModTime().Unix()}, nil
}

