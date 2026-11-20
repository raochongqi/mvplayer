package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/canvas"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	layoutBtns map[layout.Layout]*layoutButton
	layoutBox  *fyne.Container
	pauseBtn   *chipButton
	speedSel   *widget.Select
	stepSel    *widget.Select
	rateLbl    *widget.Label
	root       *fyne.Container
	statusLbl  *widget.Label
	factory    func() (player.Player, error)
	settings   Settings
	settingsPath string
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
		layoutBtns: map[layout.Layout]*layoutButton{},
		stop:       make(chan struct{}),
		factory:    player.NewDefault,
	}
	if s, err := history.Load(histPath); err == nil {
		app.hist = s
	}
	app.settingsPath = filepath.Join(filepath.Dir(histPath), "settings.json")
	app.settings = LoadSettings(app.settingsPath)
	for i := 0; i < 4; i++ {
		pane := NewPane(i, PaneOptions{
			OnRemove: app.removePane,
			OnRate:   app.setPaneRate,
			OnStep:   func(p *Pane, n int) {},
			OnToggle: app.togglePane,
			OnNudge:  app.nudgePane,
		})
		pane.SetSeekHandler(app.seekPaneAbsolute)
		pane.SetRateChoices(app.settings.rateChoices())
		pane.SetStepChoices(app.settings.stepChoices())
		app.panes = append(app.panes, pane)
	}
	app.stage.SetPanes(app.panes)
	app.stage.SetLayout(layout.One)

	app.root = container.NewBorder(app.buildToolbar(), app.buildFooter(), nil, nil, app.stage)
	w.SetContent(app.root)
	w.SetOnDropped(app.onDropped)
	w.Canvas().SetOnTypedKey(app.onKey)
	w.SetCloseIntercept(func() {
		app.persistHistory()
		w.Close()
	})
	go app.tick()
	go app.renderLoop()
	return app
}

func (a *App) buildToolbar() fyne.CanvasObject {
	openBtn := newChip("＋ 打开视频", a.openDialog)
	openBtn.SetPrimary(true)
	histBtn := newChip("≡ 历史", a.showHistory)

	a.layoutBox = container.NewHBox()
	for _, l := range layout.All {
		l := l
		b := newLayoutButton(l, func() { a.setLayout(l) })
		a.layoutBtns[l] = b
		a.layoutBox.Add(b)
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

	settingsBtn := newChip("⚙", a.showSettings)

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
	bar := container.NewBorder(nil, nil,
		container.NewCenter(left), container.NewCenter(right), container.NewCenter(center))
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
	p := newMultiPicker(a.win, func(paths []string) {
		for _, path := range paths {
			a.AddFile(path)
		}
	})
	p.Show()
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
	eng, err := a.factory()
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
	pane.pl = eng
	pane.SetFile(path)
	pane.SetRateBadge(1.0)
	pane.SetPlaying(true)
	if st, err := fileStat(path); err == nil {
		if e, ok := a.hist.Get(path, st.size, st.mtime); ok && e.PositionMs > 0 {
			pane.pendingSeek = e.PositionMs
		}
	}
	_ = eng.Play()
	a.stage.SetPanes(a.visiblePanes())
	a.relayoutForCount()
	a.refreshLayoutButtons()
}

// visiblePanes returns the filled panes in display order. The stage must only
// receive those: slicing a.panes[:n] would drop later filled panes whenever an
// earlier slot was emptied.
func (a *App) visiblePanes() []*Pane {
	out := make([]*Pane, 0, 4)
	for _, p := range a.panes {
		if p.Path != "" {
			out = append(out, p)
		}
	}
	return out
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
	if p.pl != nil {
		_ = p.pl.Close()
		a.ctrl.RemoveAt(a.playerIndex(p.pl))
		for i, pl := range a.players {
			if pl == p.pl {
				a.players = append(a.players[:i], a.players[i+1:]...)
				break
			}
		}
		p.pl = nil
	}
	p.Clear()
	a.stage.SetPanes(a.visiblePanes())
	a.relayoutForCount()
	a.refreshLayoutButtons()
}

func (a *App) playerIndex(pl player.Player) int {
	for i, q := range a.players {
		if q == pl {
			return i
		}
	}
	return -1
}

func (a *App) setPaneRate(p *Pane, rate float64) {
	p.baseRate = rate
	a.applyRate(p)
}

// applyRate pushes clamp(base+overlay) to the engine and the badge.
func (a *App) applyRate(p *Pane) {
	if p.pl == nil {
		return
	}
	eff := player.ClampRate(p.baseRate + a.overlay)
	_ = p.pl.SetRate(eff)
	p.SetRateBadge(eff)
}

func (a *App) togglePane(p *Pane) {
	if p.pl == nil {
		return
	}
	if p.pl.Paused() {
		_ = p.pl.Play()
	} else {
		_ = p.pl.Pause()
	}
}

func (a *App) nudgePane(p *Pane, ms int) {
	if p.pl == nil {
		return
	}
	target := p.pl.Position() + int64(ms)
	if target < 0 {
		target = 0
	}
	if d := p.pl.Duration(); d > 0 && target > d {
		target = d
	}
	_ = p.pl.SeekTo(target)
}

func (a *App) seekPaneAbsolute(p *Pane, f float32) {
	if p.pl == nil {
		return
	}
	if d := p.pl.Duration(); d > 0 {
		_ = p.pl.SeekTo(int64(float32(d) * f))
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
	for _, p := range a.panes {
		if p.Path != "" {
			p.baseRate = r
			a.applyRate(p)
		}
	}
	a.refreshRateLabel()
}

func (a *App) nudgeAllRates(delta float64) {
	a.overlay = clampOverlay(a.overlay + delta)
	for _, p := range a.panes {
		if p.Path != "" {
			a.applyRate(p)
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

// rateLists renders the base and effective rate lists for the status bar.
func (a *App) rateLists() (string, string) {
	var bases, effs []string
	for _, p := range a.panes {
		if p.Path == "" {
			continue
		}
		bases = append(bases, fmt.Sprintf("%.2f×", p.baseRate))
		effs = append(effs, fmt.Sprintf("%.2f×", player.ClampRate(p.baseRate+a.overlay)))
	}
	return strings.Join(bases, " / "), strings.Join(effs, " / ")
}

func (a *App) persistHistory() {
	if !a.settings.History {
		return
	}
	for _, p := range a.panes {
		if p.Path == "" || p.pl == nil {
			continue
		}
		pl := p.pl
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

func (a *App) showSettings() {
	rates := widget.NewEntry()
	rates.SetText(strings.Join(a.settings.Rates, ", "))
	steps := widget.NewEntry()
	steps.SetText(strings.Join(a.settings.Steps, ", "))
	hist := widget.NewCheck("记录播放历史", nil)
	hist.SetChecked(a.settings.History)
	form := widget.NewForm(
		widget.NewFormItem("倍速列表", rates),
		widget.NewFormItem("跳转档位", steps),
		widget.NewFormItem("历史", hist),
	)
	dialog.ShowCustomConfirm("设置", "保存", "取消", form, func(ok bool) {
		if !ok {
			return
		}
		next := Settings{
			Rates:   splitList(rates.Text),
			Steps:   splitList(steps.Text),
			History: hist.Checked,
		}
		if len(next.Rates) == 0 {
			next.Rates = defaultSettings.Rates
		}
		if len(next.Steps) == 0 {
			next.Steps = defaultSettings.Steps
		}
		a.settings = next
		_ = a.settings.Save(a.settingsPath)
		for _, p := range a.panes {
			p.SetRateChoices(next.rateChoices())
			p.SetStepChoices(next.stepChoices())
		}
	}, a.win)
}

func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '，' || r == ' ' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func (a *App) showHistory() {
	if len(a.hist.Entries) == 0 {
		dialog.ShowInformation("历史", "还没有播放记录。", a.win)
		return
	}
	entries := make([]history.Entry, 0, len(a.hist.Entries))
	for _, e := range a.hist.Entries {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].UpdatedAt > entries[j].UpdatedAt })
	list := widget.NewList(
		func() int { return len(entries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(entries) {
				o.(*widget.Label).SetText(fmt.Sprintf("%s　%s", shortName(entries[i].Path), clock(entries[i].PositionMs)))
			}
		},
	)
	var dlg *dialog.CustomDialog
	list.OnSelected = func(i widget.ListItemID) {
		list.UnselectAll()
		if i >= 0 && i < len(entries) {
			a.AddFile(entries[i].Path)
			if dlg != nil {
				dlg.Hide()
			}
		}
	}
	dlg = dialog.NewCustom("历史播放", "关闭", list, a.win)
	dlg.Resize(fyne.NewSize(640, 420))
	dlg.Show()
}

func (a *App) onKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeySpace:
		a.toggleAll()
	case fyne.KeyLeft:
		a.seekAll(-10_000)
	case fyne.KeyRight:
		a.seekAll(10_000)
	case fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5, fyne.Key6, fyne.Key7:
		idx := int(ev.Name[len(ev.Name)-1] - '1')
		if idx >= 0 && idx < len(layout.All) {
			a.setLayout(layout.All[idx])
		}
	}
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
	bases, effs := a.rateLists()
	extra := ""
	if bases != "" {
		extra = fmt.Sprintf(" · 基线 %s · 叠加后 %s", bases, effs)
	}
	a.statusLbl.SetText(fmt.Sprintf("%d 路 · %s%s", n, state, extra))
}

func (a *App) Stop() {
	a.stopOnce.Do(func() { close(a.stop) })
}

// SetFactory replaces the player constructor (tests inject fake renderers).
func (a *App) SetFactory(f func() (player.Player, error)) { a.factory = f }

func (a *App) renderLoop() {
	t := time.NewTicker(33 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-t.C:
		}
		for _, p := range a.panes {
			if p.Path == "" || p.pl == nil {
				continue
			}
			vr, ok := p.pl.(player.VideoRenderer)
			if !ok {
				continue
			}
			sz := p.Root.Size()
			scale := float32(1)
			if c := fyne.CurrentApp().Driver().CanvasForObject(p.Root); c != nil {
				scale = c.Scale()
			}
			w, h := int(sz.Width*scale), int(sz.Height*scale)
			if w < 32 || h < 32 {
				continue
			}
			if w > 1920 {
				w = 1920
			}
			if h > 1080 {
				h = 1080
			}
			if img, ok := vr.RenderFrame(w, h); ok && img != nil {
				pane := p
				fyne.Do(func() { pane.SetFrame(img) })
			}
		}
	}
}

func (a *App) tick() {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-t.C:
			a.stepOnce()
		}
	}
}

// stepOnce is the 5 Hz HUD/status pass; separated so tests can drive it.
func (a *App) stepOnce() {
	a.refreshStatus()
	for _, p := range a.panes {
		if p.Path == "" || p.pl == nil {
			continue
		}
		pl := p.pl
		if p.pendingSeek > 0 && pl.Duration() > 0 {
			_ = pl.SeekTo(p.pendingSeek)
			p.pendingSeek = 0
		}
		p.SetProgress(pl.Position(), pl.Duration())
		p.SetPlaying(!pl.Paused())
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

