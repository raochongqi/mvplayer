package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"mvplayer/internal/layout"
)

// splitter is the 5px draggable gap between panes.
type splitter struct {
	widget.BaseWidget
	vertical bool
	onDrag   func(fyne.Delta)
	rect     *canvas.Rectangle
}

func newSplitter(vertical bool, onDrag func(fyne.Delta)) *splitter {
	s := &splitter{vertical: vertical, onDrag: onDrag}
	s.rect = canvas.NewRectangle(ColorLine)
	s.ExtendBaseWidget(s)
	return s
}

func (s *splitter) Dragged(e *fyne.DragEvent) {
	if s.onDrag != nil {
		s.onDrag(e.Dragged)
	}
}

func (s *splitter) DragEnd() {}

func (s *splitter) CreateRenderer() fyne.WidgetRenderer {
	return &splitterRenderer{s: s, objects: []fyne.CanvasObject{s.rect}}
}

type splitterRenderer struct {
	s       *splitter
	objects []fyne.CanvasObject
}

func (r *splitterRenderer) Layout(size fyne.Size) { r.s.rect.Resize(size) }
func (r *splitterRenderer) MinSize() fyne.Size    { return fyne.NewSize(layout.Gap, layout.Gap) }
func (r *splitterRenderer) Refresh()              { r.s.rect.Refresh() }
func (r *splitterRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
func (r *splitterRenderer) Destroy() {}

// Stage lays the panes out and owns splitter dragging.
type Stage struct {
	widget.BaseWidget

	panes  []*Pane
	lay    layout.Layout
	ratios layout.Ratios
	vs     *splitter
	hs     *splitter
	empty  *emptySlot
}

// NewStage creates a stage with no panes.
func NewStage() *Stage {
	s := &Stage{lay: layout.One, ratios: layout.DefaultRatios()}
	s.vs = newSplitter(true, func(d fyne.Delta) { s.drag(layout.SplitVertical, d) })
	s.hs = newSplitter(false, func(d fyne.Delta) { s.drag(layout.SplitHorizontal, d) })
	s.empty = newEmptySlot()
	s.ExtendBaseWidget(s)
	return s
}

// SetPanes replaces the pane list (max 4).
func (s *Stage) SetPanes(panes []*Pane) {
	s.panes = panes
	if s.lay.PaneCount() > len(panes) {
		s.lay = layout.DefaultFor(len(panes))
	}
	s.Refresh()
}

// SetLayout switches layout, keeping current ratios.
func (s *Stage) SetLayout(l layout.Layout) {
	s.lay = l
	s.Refresh()
}

// Layout returns the current layout.
func (s *Stage) Layout() layout.Layout { return s.lay }

// Ratios returns the current split ratios.
func (s *Stage) Ratios() layout.Ratios { return s.ratios }

// ResetRatios restores an even split (double-click on a splitter in the UI).
func (s *Stage) ResetRatios() {
	s.ratios = layout.DefaultRatios()
	s.Refresh()
}

func (s *Stage) drag(which layout.Splitter, d fyne.Delta) {
	size := s.Size()
	switch which {
	case layout.SplitVertical:
		if size.Width-layout.Gap > 0 {
			s.ratios.X = layout.Drag(s.lay, size.Width, size.Height, s.ratios, which,
				(s.Size().Width-layout.Gap)*s.ratios.X+d.DX, 0).X
		}
	case layout.SplitHorizontal:
		if size.Height-layout.Gap > 0 {
			s.ratios.Y = layout.Drag(s.lay, size.Width, size.Height, s.ratios, which,
				0, (size.Height-layout.Gap)*s.ratios.Y+d.DY).Y
		}
	}
	s.Refresh()
}

func (s *Stage) CreateRenderer() fyne.WidgetRenderer {
	r := &stageRenderer{stage: s, bg: canvas.NewRectangle(ColorBG)}
	r.rebuild()
	return r
}

type stageRenderer struct {
	stage   *Stage
	bg      *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *stageRenderer) rebuild() {
	r.objects = []fyne.CanvasObject{r.bg}
	for _, p := range r.stage.panes {
		r.objects = append(r.objects, p.Root)
	}
	r.objects = append(r.objects, r.stage.vs, r.stage.hs, r.stage.empty)
}

func hasV(l layout.Layout) bool {
	switch l {
	case layout.TwoV, layout.ThreeL, layout.ThreeR, layout.ThreeQ, layout.Four:
		return true
	}
	return false
}

func hasH(l layout.Layout) bool {
	switch l {
	case layout.TwoH, layout.ThreeQ:
		return true
	}
	return false
}

func (r *stageRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	if len(r.objects) != 1+len(r.stage.panes)+3 {
		r.rebuild()
	}
	rects := layout.Compute(r.stage.lay, size.Width, size.Height, r.stage.ratios, layout.Gap)
	for i, p := range r.stage.panes {
		if i < len(rects) {
			p.Root.Move(fyne.NewPos(rects[i].X, rects[i].Y))
			p.Root.Resize(fyne.NewSize(rects[i].W, rects[i].H))
			p.Root.Show()
		} else {
			p.Root.Hide()
		}
	}
	// empty slot: whole area when nothing is open, otherwise the spare cell
	if len(r.stage.panes) == 0 {
		r.stage.empty.Move(fyne.NewPos(0, 0))
		r.stage.empty.Resize(size)
		r.stage.empty.Show()
	} else if cell, ok := layout.EmptyCell(r.stage.lay, size.Width, size.Height, r.stage.ratios, layout.Gap); ok {
		r.stage.empty.Move(fyne.NewPos(cell.X, cell.Y))
		r.stage.empty.Resize(fyne.NewSize(cell.W, cell.H))
		r.stage.empty.Show()
	} else if len(rects) > len(r.stage.panes) {
		cell := rects[len(r.stage.panes)]
		r.stage.empty.Move(fyne.NewPos(cell.X, cell.Y))
		r.stage.empty.Resize(fyne.NewSize(cell.W, cell.H))
		r.stage.empty.Show()
	} else {
		r.stage.empty.Hide()
	}
	w, h := size.Width, size.Height
	if hasV(r.stage.lay) {
		x := (w - layout.Gap) * r.stage.ratios.X
		r.stage.vs.Move(fyne.NewPos(x, 0))
		r.stage.vs.Resize(fyne.NewSize(layout.Gap, h))
		r.stage.vs.Show()
	} else {
		r.stage.vs.Hide()
	}
	if hasH(r.stage.lay) {
		y := (h - layout.Gap) * r.stage.ratios.Y
		r.stage.hs.Move(fyne.NewPos(0, y))
		r.stage.hs.Resize(fyne.NewSize(w, layout.Gap))
		r.stage.hs.Show()
	} else {
		r.stage.hs.Hide()
	}
}

func (r *stageRenderer) MinSize() fyne.Size { return fyne.NewSize(320, 180) }

func (r *stageRenderer) Refresh() {
	r.bg.FillColor = ColorBG
	r.bg.Refresh()
	r.Layout(r.stage.Size())
	for _, p := range r.stage.panes {
		p.Root.Refresh()
	}
	canvas.Refresh(r.stage)
}

func (r *stageRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *stageRenderer) Destroy()                     {}
