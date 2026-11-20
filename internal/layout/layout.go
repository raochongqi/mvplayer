// Package layout computes pane rectangles for the six supported layouts and
// owns the splitter drag math. It is pure Go so it can be tested headless.
package layout

// Layout identifies one of the supported pane arrangements.
type Layout string

const (
	One    Layout = "1"   // single pane
	TwoV   Layout = "2v"  // left | right
	TwoH   Layout = "2h"  // top  / bottom
	ThreeL Layout = "3lt" // left column + right top + right bottom
	ThreeR Layout = "3rt" // right column + left top + left bottom
	ThreeQ Layout = "3q"  // 2x2 grid with the bottom-right cell empty
	Four   Layout = "4"   // 2x2 grid
)

// All lists layouts in toolbar order.
var All = []Layout{One, TwoV, TwoH, ThreeL, ThreeR, ThreeQ, Four}

// PaneCount reports how many video panes a layout shows.
func (l Layout) PaneCount() int {
	switch l {
	case One:
		return 1
	case TwoV, TwoH:
		return 2
	case ThreeL, ThreeR, ThreeQ:
		return 3
	case Four:
		return 4
	}
	return 0
}

// ValidFor reports whether a layout makes sense for n videos (1..4).
func ValidFor(l Layout, n int) bool { return l.PaneCount() == n }

// DefaultFor returns the layout selected automatically after opening n videos.
func DefaultFor(n int) Layout {
	switch n {
	case 1:
		return One
	case 2:
		return TwoV
	case 3:
		return ThreeL
	default:
		return Four
	}
}

// Rect is a pane rectangle in window coordinates.
type Rect struct{ X, Y, W, H float32 }

// Ratios holds the two draggable split positions, both in (0,1).
type Ratios struct{ X, Y float32 }

// DefaultRatios is an even split.
func DefaultRatios() Ratios { return Ratios{X: 0.5, Y: 0.5} }

// Gap is the visual (and draggable) space between panes, in pixels.
const Gap float32 = 5

// MinRatio/MaxRatio clamp splitter drags so a pane can never vanish.
const (
	MinRatio float32 = 0.1
	MaxRatio float32 = 0.9
)

// Clamp keeps split ratios inside the allowed range.
func (r Ratios) Clamp() Ratios {
	return Ratios{X: clamp(r.X), Y: clamp(r.Y)}
}

func clamp(v float32) float32 {
	if v < MinRatio {
		return MinRatio
	}
	if v > MaxRatio {
		return MaxRatio
	}
	return v
}

// Compute returns one Rect per pane, in fill order (a, b, c, d).
//
// w/h are the content area in pixels; r is the split ratio; gap is the space
// between panes. Panes are laid out so that the whole content area is used.
func Compute(l Layout, w, h float32, r Ratios, gap float32) []Rect {
	r = r.Clamp()
	innerW := w - gap
	innerH := h - gap
	lw := innerW * r.X
	rw := innerW - lw
	th := innerH * r.Y
	bh := innerH - th

	switch l {
	case One:
		return []Rect{{0, 0, w, h}}
	case TwoV:
		return []Rect{{0, 0, lw, h}, {lw + gap, 0, rw, h}}
	case TwoH:
		return []Rect{{0, 0, w, th}, {0, th + gap, w, bh}}
	case ThreeL:
		return []Rect{
			{0, 0, lw, h},
			{lw + gap, 0, rw, th},
			{lw + gap, th + gap, rw, bh},
		}
	case ThreeR:
		return []Rect{
			{0, 0, lw, th},
			{0, th + gap, lw, bh},
			{lw + gap, 0, rw, h},
		}
	case ThreeQ:
		return []Rect{
			{0, 0, lw, th},
			{lw + gap, 0, rw, th},
			{0, th + gap, lw, bh},
		}
	case Four:
		return []Rect{
			{0, 0, lw, th},
			{lw + gap, 0, rw, th},
			{0, th + gap, lw, bh},
			{lw + gap, th + gap, rw, bh},
		}
	}
	return nil
}

// EmptyCell returns the rectangle of the unused cell in ThreeQ (the
// bottom-right quadrant). Other layouts have no empty cell.
func EmptyCell(l Layout, w, h float32, r Ratios, gap float32) (Rect, bool) {
	if l != ThreeQ {
		return Rect{}, false
	}
	r = r.Clamp()
	innerW := w - gap
	innerH := h - gap
	x := innerW*r.X + gap
	y := innerH*r.Y + gap
	return Rect{x, y, innerW - innerW*r.X, innerH - innerH*r.Y}, true
}

// Splitter identifies a draggable divider.
type Splitter int

const (
	SplitNone Splitter = iota
	SplitVertical
	SplitHorizontal
)

// HitTest reports which splitter (if any) is under the point, within tol px.
//
// A layout only exposes the splitters it actually draws: a single pane has
// none, a two-column layout has one vertical splitter, and so on.
func HitTest(l Layout, w, h float32, r Ratios, x, y float32, tol float32) Splitter {
	r = r.Clamp()
	hasV := false
	hasH := false
	switch l {
	case TwoV, ThreeL, ThreeR, Four:
		hasV = true
	case TwoH:
		hasH = true
	case ThreeQ:
		hasV, hasH = true, true
	}
	if hasV {
		line := (w - Gap) * r.X
		if x >= line-tol && x <= line+Gap+tol {
			return SplitVertical
		}
	}
	if hasH {
		line := (h - Gap) * r.Y
		if y >= line-tol && y <= line+Gap+tol {
			return SplitHorizontal
		}
	}
	return SplitNone
}

// Drag moves a splitter to an absolute pointer coordinate, clamped.
func Drag(l Layout, w, h float32, r Ratios, s Splitter, x, y float32) Ratios {
	r = r.Clamp()
	switch s {
	case SplitVertical:
		if w-Gap > 0 {
			r.X = clamp(x / (w - Gap))
		}
	case SplitHorizontal:
		if h-Gap > 0 {
			r.Y = clamp(y / (h - Gap))
		}
	}
	return r
}

// ToggleQQEmpty reports the empty cell of ThreeQ in (row, col) form; the
// bottom-right cell stays empty by design.
func ToggleQQEmpty() (row, col int) { return 1, 1 }
