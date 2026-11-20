package layout

import (
	"math"
	"testing"
)

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 0.01 }

func TestPaneCounts(t *testing.T) {
	want := map[Layout]int{One: 1, TwoV: 2, TwoH: 2, ThreeL: 3, ThreeR: 3, ThreeQ: 3, Four: 4}
	for l, n := range want {
		if got := l.PaneCount(); got != n {
			t.Fatalf("%s: PaneCount=%d want %d", l, got, n)
		}
	}
	if len(All) != 7 {
		t.Fatalf("All has %d layouts, want 7", len(All))
	}
}

func TestValidFor(t *testing.T) {
	cases := []struct {
		l Layout
		n int
		ok bool
	}{
		{One, 1, true}, {One, 2, false},
		{TwoV, 2, true}, {TwoH, 2, true}, {TwoV, 3, false},
		{ThreeL, 3, true}, {ThreeR, 3, true}, {ThreeQ, 3, true}, {ThreeQ, 4, false},
		{Four, 4, true}, {Four, 3, false},
	}
	for _, c := range cases {
		if got := ValidFor(c.l, c.n); got != c.ok {
			t.Fatalf("ValidFor(%s,%d)=%v want %v", c.l, c.n, got, c.ok)
		}
	}
}

func TestComputeFillsAreaWithoutOverlap(t *testing.T) {
	const w, h = 1440, 850
	for _, l := range All {
		rects := Compute(l, w, h, DefaultRatios(), Gap)
		if len(rects) != l.PaneCount() {
			t.Fatalf("%s: %d rects, want %d", l, len(rects), l.PaneCount())
		}
		var area float64
		for i, r := range rects {
			if r.W <= 0 || r.H <= 0 {
				t.Fatalf("%s pane %d has non-positive size %+v", l, i, r)
			}
			if r.X < 0 || r.Y < 0 || r.X+r.W > w+0.01 || r.Y+r.H > h+0.01 {
				t.Fatalf("%s pane %d out of bounds: %+v", l, i, r)
			}
			area += float64(r.W) * float64(r.H)
			for j := 0; j < i; j++ {
				if overlaps(rects[j], r) {
					t.Fatalf("%s panes %d and %d overlap: %+v %+v", l, j, i, rects[j], r)
				}
			}
		}
		// Panes plus inter-pane gaps must account for the whole area.
		gaps := float64(w)*float64(Gap) + float64(h)*float64(Gap) - float64(Gap)*float64(Gap)
		lw := float64(w-Gap) * float64(DefaultRatios().X)
		rw := float64(w-Gap) - lw
		switch l {
		case One:
			gaps = 0
		case TwoV:
			gaps = float64(h) * float64(Gap)
		case TwoH:
			gaps = float64(w) * float64(Gap)
		case ThreeL:
			gaps = float64(h)*float64(Gap) + rw*float64(Gap)
		case ThreeR:
			gaps = float64(h)*float64(Gap) + lw*float64(Gap)
		case ThreeQ: // horizontal gap runs under the empty cell too
			gaps = float64(h)*float64(Gap) + float64(w)*float64(Gap) - float64(Gap)*float64(Gap)
		case Four:
			gaps = float64(w)*float64(Gap) + float64(h)*float64(Gap) - float64(Gap)*float64(Gap)
		}
		empty := 0.0
		if l == ThreeQ { // bottom-right cell is a drop target, not a video pane
			th := float64(h-Gap) * float64(DefaultRatios().Y)
			bh := float64(h-Gap) - th
			empty = rw * bh
		}
		if math.Abs(area+float64(gaps)+empty-float64(w*h)) > 1 {
			t.Fatalf("%s: area %.1f + gaps %.1f + empty %.1f != %d", l, area, gaps, empty, w*h)
		}
	}
}

func overlaps(a, b Rect) bool {
	return a.X < b.X+b.W-0.01 && b.X < a.X+a.W-0.01 && a.Y < b.Y+b.H-0.01 && b.Y < a.Y+a.H-0.01
}

func TestClampRatios(t *testing.T) {
	for _, in := range []Ratios{{0, 0}, {1, 1}, {-5, 2}} {
		got := in.Clamp()
		if got.X < MinRatio || got.X > MaxRatio || got.Y < MinRatio || got.Y > MaxRatio {
			t.Fatalf("Clamp(%+v)=%+v out of range", in, got)
		}
	}
}

func TestHitTest(t *testing.T) {
	const w, h = 1000, 600
	r := DefaultRatios()
	// vertical splitter of TwoV sits at (w-gap)*0.5
	if got := HitTest(TwoV, w, h, r, (w-Gap)*0.5+2, 300, 6); got != SplitVertical {
		t.Fatalf("TwoV hit=%v want vertical", got)
	}
	if got := HitTest(TwoH, w, h, r, 500, (h-Gap)*0.5+2, 6); got != SplitHorizontal {
		t.Fatalf("TwoH hit=%v want horizontal", got)
	}
	if got := HitTest(One, w, h, r, 500, 300, 6); got != SplitNone {
		t.Fatalf("One hit=%v want none", got)
	}
	if got := HitTest(Four, w, h, r, (w-Gap)*0.5+2, (h-Gap)*0.5+2, 6); got != SplitVertical && got != SplitHorizontal {
		t.Fatalf("Four corner hit=%v want a splitter", got)
	}
	if got := HitTest(ThreeQ, w, h, r, 500, 300, 6); got != SplitVertical {
		t.Fatalf("ThreeQ center hit=%v want vertical (first match)", got)
	}
}

func TestDragClamps(t *testing.T) {
	r := Drag(TwoV, 1000, 600, DefaultRatios(), SplitVertical, -50, 0)
	if r.X != MinRatio {
		t.Fatalf("drag left clamp=%v want %v", r.X, MinRatio)
	}
	r = Drag(TwoV, 1000, 600, DefaultRatios(), SplitVertical, 5000, 0)
	if r.X != MaxRatio {
		t.Fatalf("drag right clamp=%v want %v", r.X, MaxRatio)
	}
	r = Drag(TwoV, 1000, 600, DefaultRatios(), SplitNone, 800, 0)
	if r.X != 0.5 || r.Y != 0.5 {
		t.Fatalf("drag on no splitter changed ratios: %+v", r)
	}
}

func TestDefaultFor(t *testing.T) {
	for n, want := range map[int]Layout{1: One, 2: TwoV, 3: ThreeL, 4: Four} {
		if got := DefaultFor(n); got != want {
			t.Fatalf("DefaultFor(%d)=%s want %s", n, got, want)
		}
	}
}
