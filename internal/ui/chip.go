package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// chipButton is the flat 1px-outlined control used by the approved design.
// Fyne's stock button draws a filled grey block, which is what made the first
// screenshots look different from the design.
type chipButton struct {
	widget.BaseWidget

	text    string
	active  bool
	primary bool
	enabled bool
	hover   bool
	onTap   func()

	bg     *canvas.Rectangle
	border *canvas.Rectangle
	label  *canvas.Text
}

func newChip(text string, onTap func()) *chipButton {
	c := &chipButton{text: text, enabled: true, onTap: onTap}
	c.label = canvas.NewText(text, ColorText)
	c.label.TextSize = 12
	c.label.Alignment = fyne.TextAlignCenter
	c.bg = canvas.NewRectangle(color.Transparent)
	c.bg.CornerRadius = 3
	c.border = canvas.NewRectangle(color.Transparent)
	c.border.CornerRadius = 3
	c.border.StrokeColor = ColorLine
	c.border.StrokeWidth = 1
	c.ExtendBaseWidget(c)
	return c
}

// SetPrimary draws the accent outline (used by the open button).
func (c *chipButton) SetPrimary(on bool) { c.primary = on; c.Refresh() }

// SetActive marks the current layout.
func (c *chipButton) SetActive(on bool) { c.active = on; c.Refresh() }

// SetEnabled toggles the dimmed state.
func (c *chipButton) SetEnabled(on bool) { c.enabled = on; c.Refresh() }

// Disabled reports the dimmed state (used by tests).
func (c *chipButton) Disabled() bool { return !c.enabled }

// SetText changes the label.
func (c *chipButton) SetText(s string) { c.text = s; c.label.Text = s; c.Refresh() }

func (c *chipButton) Tapped(*fyne.PointEvent) {
	if c.enabled && c.onTap != nil {
		c.onTap()
	}
}

func (c *chipButton) MouseIn(*fyne.PointEvent) { c.hover = true; c.Refresh() }
func (c *chipButton) MouseOut()               { c.hover = false; c.Refresh() }

func (c *chipButton) CreateRenderer() fyne.WidgetRenderer {
	return &chipRenderer{c: c, objects: []fyne.CanvasObject{c.bg, c.border, c.label}}
}

type chipRenderer struct {
	c       *chipButton
	objects []fyne.CanvasObject
}

func (r *chipRenderer) Layout(size fyne.Size) {
	r.c.bg.Resize(size)
	r.c.border.Resize(size)
	r.c.label.Resize(size)
	// canvas.Text draws from its top edge, so centre by the measured height
	// instead of a hard-coded baseline.
	h := r.c.label.MinSize().Height
	r.c.label.Move(fyne.NewPos(0, (size.Height-h)/2))
}

func (r *chipRenderer) MinSize() fyne.Size {
	w := r.c.label.MinSize().Width + 14
	if w < 26 {
		w = 26
	}
	return fyne.NewSize(w, 26)
}

func (r *chipRenderer) Refresh() {
	c := r.c
	switch {
	case !c.enabled:
		c.label.Color = ColorMuted
		c.border.StrokeColor = ColorLine
		c.bg.FillColor = color.Transparent
	case c.primary:
		c.label.Color = ColorAccent
		c.border.StrokeColor = ColorAccent
		c.bg.FillColor = ColorAccentWeak
	case c.active:
		c.label.Color = ColorText
		c.border.StrokeColor = ColorAccent
		c.bg.FillColor = ColorAccentWeak
	case c.hover:
		c.label.Color = ColorText
		c.border.StrokeColor = ColorLine
		c.bg.FillColor = ColorAccentWeak
	default:
		c.label.Color = ColorText
		c.border.StrokeColor = ColorLine
		c.bg.FillColor = color.Transparent
	}
	c.bg.Refresh()
	c.border.Refresh()
	c.label.Refresh()
}

func (r *chipRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *chipRenderer) Destroy()                     {}
