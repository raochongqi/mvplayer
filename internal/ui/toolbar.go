package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// toolbar lays its three groups out on one strict horizontal line: every item
// is centred vertically at (height-itemHeight)/2, the left group hugs the left
// edge, the right group the right edge and the middle group is centred on the
// window. Fyne's HBox top-aligns children, which is what made the toolbar look
// crooked.
type toolbar struct {
	widget.BaseWidget

	left  []fyne.CanvasObject
	mid   []fyne.CanvasObject
	right []fyne.CanvasObject

	bg *canvas.Rectangle
}

// NewToolbar builds the bar; items keep their own MinSize, so wrap fixed-width
// widgets (selects) in a GridWrap before passing them in.
func NewToolbar(left, mid, right []fyne.CanvasObject) *toolbar {
	t := &toolbar{left: left, mid: mid, right: right}
	t.bg = canvas.NewRectangle(ColorChrome)
	t.ExtendBaseWidget(t)
	return t
}

const (
	toolbarHeight float32 = 44
	toolbarPad    float32 = 12
	toolbarGap    float32 = 8
)

func (t *toolbar) CreateRenderer() fyne.WidgetRenderer {
	objs := []fyne.CanvasObject{t.bg}
	objs = append(objs, t.left...)
	objs = append(objs, t.mid...)
	objs = append(objs, t.right...)
	return &toolbarRenderer{t: t, objects: objs}
}

type toolbarRenderer struct {
	t       *toolbar
	objects []fyne.CanvasObject
}

func groupWidth(items []fyne.CanvasObject) (w, h float32) {
	for i, o := range items {
		m := o.MinSize()
		if i > 0 {
			w += toolbarGap
		}
		w += m.Width
		if m.Height > h {
			h = m.Height
		}
	}
	return w, h
}

func (r *toolbarRenderer) Layout(size fyne.Size) {
	r.t.bg.Resize(size)
	midW, _ := groupWidth(r.t.mid)
	// middle group centred on the window
	x := (size.Width - midW) / 2
	for i, o := range r.t.mid {
		m := o.MinSize()
		if i > 0 {
			x += toolbarGap
		}
		o.Resize(fyne.NewSize(m.Width, m.Height))
		o.Move(fyne.NewPos(x, (size.Height-m.Height)/2))
		x += m.Width
	}
	// left group
	x = toolbarPad
	for i, o := range r.t.left {
		m := o.MinSize()
		if i > 0 {
			x += toolbarGap
		}
		o.Resize(fyne.NewSize(m.Width, m.Height))
		o.Move(fyne.NewPos(x, (size.Height-m.Height)/2))
		x += m.Width
	}
	// right group, laid out right-to-left
	x = size.Width - toolbarPad
	for i := len(r.t.right) - 1; i >= 0; i-- {
		o := r.t.right[i]
		m := o.MinSize()
		x -= m.Width
		o.Resize(fyne.NewSize(m.Width, m.Height))
		o.Move(fyne.NewPos(x, (size.Height-m.Height)/2))
		x -= toolbarGap
	}
}

func (r *toolbarRenderer) MinSize() fyne.Size { return fyne.NewSize(320, toolbarHeight) }

func (r *toolbarRenderer) Refresh() {
	r.t.bg.FillColor = ColorChrome
	r.t.bg.Refresh()
	for _, o := range r.objects {
		o.Refresh()
	}
}

func (r *toolbarRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *toolbarRenderer) Destroy()                     {}
