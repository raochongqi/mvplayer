package ui

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SeekBar is the thin draggable progress bar drawn above the pane HUD.
type SeekBar struct {
	widget.BaseWidget
	fraction float32
	onSeek   func(float32)
}

// NewSeekBar creates a bar at 0.
func NewSeekBar(onSeek func(float32)) *SeekBar {
	s := &SeekBar{onSeek: onSeek}
	s.ExtendBaseWidget(s)
	return s
}

// SetFraction moves the fill without firing onSeek.
func (s *SeekBar) SetFraction(f float32) {
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	if s.fraction == f {
		return
	}
	s.fraction = f
	s.Refresh()
}

// Dragged maps the pointer to an absolute fraction.
func (s *SeekBar) Dragged(e *fyne.DragEvent) {
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(s)
	if w := s.Size().Width; w > 0 {
		s.SetFraction((e.Position.X - pos.X) / w)
	}
}

// DragEnd commits the seek.
func (s *SeekBar) DragEnd() {
	if s.onSeek != nil {
		s.onSeek(s.fraction)
	}
}

// Tapped also seeks (single click).
func (s *SeekBar) Tapped(e *fyne.PointEvent) {
	if w := s.Size().Width; w > 0 {
		s.SetFraction(e.Position.X / w)
	}
	if s.onSeek != nil {
		s.onSeek(s.fraction)
	}
}

func (s *SeekBar) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(ColorLine)
	fill := canvas.NewRectangle(ColorAccent)
	return &seekBarRenderer{s: s, track: track, fill: fill, objects: []fyne.CanvasObject{track, fill}}
}

type seekBarRenderer struct {
	s       *SeekBar
	track   *canvas.Rectangle
	fill    *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *seekBarRenderer) Layout(size fyne.Size) {
	const h float32 = 3
	y := (size.Height - h) / 2
	r.track.Move(fyne.NewPos(0, y))
	r.track.Resize(fyne.NewSize(size.Width, h))
	r.fill.Move(fyne.NewPos(0, y))
	r.fill.Resize(fyne.NewSize(size.Width*r.s.fraction, h))
}

func (r *seekBarRenderer) MinSize() fyne.Size { return fyne.NewSize(40, 14) }
func (r *seekBarRenderer) Refresh() {
	r.fill.Resize(fyne.NewSize(r.s.Size().Width*r.s.fraction, 3))
	r.track.Refresh()
	r.fill.Refresh()
}
func (r *seekBarRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *seekBarRenderer) Destroy()                     {}

// Pane is one video cell: surface, file name, rate badge and HUD.
// It deliberately does not embed the container: Fyne renders embedded
// containers as opaque objects, so the stage positions Root instead.
type Pane struct {
	Root *fyne.Container

	Index int
	Path  string

	title    *widget.Label
	badge    *widget.Label
	timeLbl  *widget.Label
	seek     *SeekBar
	rateSel  *widget.Select
	stepSel  *widget.Select
	playBtn  *chipButton
	surface  *canvas.Rectangle
	image    *canvas.Image
	onRemove func(*Pane)
	onRate   func(*Pane, float64)
	onStep   func(*Pane, int)
	onToggle func(*Pane)
	onNudge  func(*Pane, int)
	onSeekAt func(*Pane, float32)

	stepSeconds int
	baseRate    float64
	pendingSeek int64
	playing     bool
}

// PaneOptions wires pane callbacks.
type PaneOptions struct {
	OnRemove func(*Pane)
	OnRate   func(*Pane, float64)
	OnStep   func(*Pane, int)
	OnToggle func(*Pane)
	OnNudge  func(*Pane, int)
}

// NewPane builds an empty pane.
func NewPane(index int, opts PaneOptions) *Pane {
	p := &Pane{
		Index:       index,
		stepSeconds: 10,
		baseRate:    1,
		onRemove:    opts.OnRemove,
		onRate:      opts.OnRate,
		onStep:      opts.OnStep,
		onToggle:    opts.OnToggle,
		onNudge:     opts.OnNudge,
	}
	p.title = widget.NewLabel("空位 · 拖入视频")
	p.title.TextStyle = fyne.TextStyle{Bold: false}
	p.badge = widget.NewLabel("1.0×")
	p.badge.TextStyle = fyne.TextStyle{Bold: true}
	p.timeLbl = widget.NewLabel("00:00:00 / 00:00:00")
	p.seek = NewSeekBar(func(f float32) { p.seekTo(f) })

	p.rateSel = widget.NewSelect(presetRates(), func(string) {})
	p.rateSel.SetSelected("1.0×")
	p.rateSel.OnChanged = func(s string) {
		if p.onRate != nil {
			if r, ok := parseRate(s); ok {
				p.onRate(p, r)
			}
		}
	}
	p.stepSel = widget.NewSelect([]string{"10s", "30s", "60s", "120s"}, func(s string) {
		if n, ok := parseStep(s); ok {
			p.stepSeconds = n
			if p.onStep != nil {
				p.onStep(p, n)
			}
		}
	})
	p.stepSel.SetSelected("10s")

	p.playBtn = newChip("Ⅱ", func() { p.toggle() })
	minBtn := newChip("−", func() { p.nudge(-1) })
	plusBtn := newChip("＋", func() { p.nudge(1) })
	closeBtn := newChip("✕", func() {
		if p.onRemove != nil {
			p.onRemove(p)
		}
	})

	p.surface = canvas.NewRectangle(ColorPane)
	p.image = canvas.NewImageFromResource(nil)
	p.image.FillMode = canvas.ImageFillContain
	p.image.Hide()

	p.badge.TextStyle = fyne.TextStyle{Bold: true}

	topRow := container.NewHBox(chip(p.title, ColorChipBG), layoutSpacer(), chip(p.badge, ColorAccentWeak))
	stepBox := container.NewGridWrap(fyne.NewSize(56, 26), p.stepSel)
	rateBox := container.NewGridWrap(fyne.NewSize(62, 26), p.rateSel)
	hudRight := container.NewHBox(stepBox, minBtn, p.playBtn, plusBtn, rateBox, closeBtn)
	hud := container.NewBorder(nil, nil, container.NewCenter(p.timeLbl), hudRight, p.seek)
	hudBox := container.NewStack(canvas.NewRectangle(ColorHUD), container.NewPadded(hud))
	overlay := container.NewBorder(topRow, hudBox, nil, nil, nil)
	p.Root = container.NewStack(p.surface, p.image, overlay)
	return p
}

// SetRateChoices replaces the speed dropdown entries.
func (p *Pane) SetRateChoices(list []string) {
	p.rateSel.Options = list
	p.rateSel.Refresh()
}

// SetStepChoices replaces the seek-step dropdown entries.
func (p *Pane) SetStepChoices(list []string) {
	p.stepSel.Options = list
	p.stepSel.Refresh()
}

// SetFile shows a file in this pane.
func (p *Pane) SetFile(path string) {
	p.Path = path
	p.title.SetText(shortName(path))
	p.title.Refresh()
}

// Clear empties the pane.
func (p *Pane) Clear() {
	p.Path = ""
	p.title.SetText("空位 · 拖入视频")
	p.badge.SetText("1.0×")
	p.timeLbl.SetText("00:00:00 / 00:00:00")
	p.seek.SetFraction(0)
	p.image.Hide()
	p.playing = false
	p.playBtn.SetText("▶")
}

// SetProgress updates HUD time and bar (pos/dur in ms).
func (p *Pane) SetProgress(pos, dur int64) {
	p.timeLbl.SetText(fmt.Sprintf("%s / %s", clock(pos), clock(dur)))
	if dur > 0 {
		p.seek.SetFraction(float32(pos) / float32(dur))
	}
}

// SetRateBadge mirrors the effective rate.
func (p *Pane) SetRateBadge(r float64) {
	p.badge.SetText(fmt.Sprintf("%.2f×", r))
}

// SetPlaying updates the play glyph.
func (p *Pane) SetPlaying(playing bool) {
	p.playing = playing
	if playing {
		p.playBtn.SetText("Ⅱ")
	} else {
		p.playBtn.SetText("▶")
	}
}

// Image exposes the frame surface for the renderer.
func (p *Pane) Image() *canvas.Image { return p.image }

// SetFrame paints one decoded video frame.
func (p *Pane) SetFrame(img image.Image) {
	if img == nil {
		return
	}
	p.image.Image = img
	if !p.image.Visible() {
		p.image.Show()
	}
	p.image.Refresh()
}

func (p *Pane) toggle() {
	if p.onToggle != nil {
		p.onToggle(p)
	}
}

func (p *Pane) nudge(dir int) {
	if p.onNudge != nil {
		p.onNudge(p, dir*p.stepSeconds)
	}
}

func (p *Pane) seekTo(f float32) {
	if p.onSeekAt != nil {
		p.onSeekAt(p, f)
	}
}

// OnSeekAt is set by the app to perform an absolute seek.
func (p *Pane) SetSeekHandler(fn func(*Pane, float32)) { p.onSeekAt = fn }

