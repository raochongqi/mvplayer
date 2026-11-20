package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"mvplayer/internal/layout"
)

// layoutButton draws the approved-design grid icon instead of a glyph, so the
// toolbar matches approved.png.
type layoutButton struct {
	widget.BaseWidget

	kind    layout.Layout
	active  bool
	enabled bool
	hover   bool
	onTap   func()

	cell  [4]*canvas.Rectangle
	slots [][4]float32 // x,y,w,h per cell in a 14x10 box
}

func newLayoutButton(kind layout.Layout, onTap func()) *layoutButton {
	b := &layoutButton{kind: kind, enabled: true, onTap: onTap}
	for i := range b.cell {
		b.cell[i] = canvas.NewRectangle(ColorMuted)
		b.cell[i].CornerRadius = 0.5
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *layoutButton) SetActive(on bool) { b.active = on; b.Refresh() }
func (b *layoutButton) SetEnabled(on bool) { b.enabled = on; b.Refresh() }
func (b *layoutButton) Disabled() bool     { return !b.enabled }
func (b *layoutButton) Tapped(*fyne.PointEvent) {
	if b.enabled && b.onTap != nil {
		b.onTap()
	}
}
func (b *layoutButton) MouseIn(*fyne.PointEvent) { b.hover = true; b.Refresh() }
func (b *layoutButton) MouseOut()               { b.hover = false; b.Refresh() }

func (b *layoutButton) CreateRenderer() fyne.WidgetRenderer {
	r := &layoutButtonRenderer{b: b}
	r.objects = []fyne.CanvasObject{}
	for i := range b.cell {
		r.objects = append(r.objects, b.cell[i])
	}
	return r
}

type layoutButtonRenderer struct {
	b       *layoutButton
	objects []fyne.CanvasObject
}

// cells returns the icon rectangles for a layout in a 14x10 box.
func cells(l layout.Layout) [][4]float32 {
	switch l {
	case layout.One:
		return [][4]float32{{0, 0, 14, 10}}
	case layout.TwoV:
		return [][4]float32{{0, 0, 6.5, 10}, {7.5, 0, 6.5, 10}}
	case layout.TwoH:
		return [][4]float32{{0, 0, 14, 4.5}, {0, 5.5, 14, 4.5}}
	case layout.ThreeL:
		return [][4]float32{{0, 0, 6.5, 10}, {7.5, 0, 6.5, 4.5}, {7.5, 5.5, 6.5, 4.5}}
	case layout.ThreeR:
		return [][4]float32{{0, 0, 6.5, 4.5}, {0, 5.5, 6.5, 4.5}, {7.5, 0, 6.5, 10}}
	case layout.ThreeQ:
		return [][4]float32{{0, 0, 6.5, 4.5}, {7.5, 0, 6.5, 4.5}, {0, 5.5, 6.5, 4.5}}
	case layout.Four:
		return [][4]float32{{0, 0, 6.5, 4.5}, {7.5, 0, 6.5, 4.5}, {0, 5.5, 6.5, 4.5}, {7.5, 5.5, 6.5, 4.5}}
	}
	return nil
}

func (r *layoutButtonRenderer) Layout(size fyne.Size) {
	cs := cells(r.b.kind)
	sx := (size.Width - 14) / 2
	sy := (size.Height - 10) / 2
	for i := range r.b.cell {
		if i < len(cs) {
			c := cs[i]
			r.b.cell[i].Move(fyne.NewPos(sx+c[0], sy+c[1]))
			r.b.cell[i].Resize(fyne.NewSize(c[2], c[3]))
			r.b.cell[i].Show()
		} else {
			r.b.cell[i].Hide()
		}
	}
}

func (r *layoutButtonRenderer) MinSize() fyne.Size { return fyne.NewSize(30, 26) }

func (r *layoutButtonRenderer) Refresh() {
	col := ColorMuted
	switch {
	case !r.b.enabled:
		col = color.NRGBA{R: 0x4a, G: 0x4e, B: 0x54, A: 0xff}
	case r.b.active:
		col = ColorAccent
	case r.b.hover:
		col = ColorText
	}
	for i := range r.b.cell {
		r.b.cell[i].FillColor = col
		r.b.cell[i].Refresh()
	}
}

func (r *layoutButtonRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *layoutButtonRenderer) Destroy()                     {}
