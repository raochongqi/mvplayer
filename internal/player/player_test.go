package player

import (
	"errors"
	"testing"
)

type fake struct {
	paused   bool
	pos, dur int64
	rate     float64
	failOn   string
}

func (f *fake) Open(string) error  { return nil }
func (f *fake) Close() error       { return nil }
func (f *fake) Play() error        { f.paused = false; return f.err("play") }
func (f *fake) Pause() error       { f.paused = true; return f.err("pause") }
func (f *fake) Paused() bool       { return f.paused }
func (f *fake) Position() int64    { return f.pos }
func (f *fake) Duration() int64    { return f.dur }
func (f *fake) Rate() float64      { return f.rate }
func (f *fake) Seek(ms int64) error {
	if ms < 0 {
		ms = 0
	}
	f.pos = ms
	return f.err("seek")
}
func (f *fake) SetRate(r float64) error { f.rate = r; return f.err("rate") }
func (f *fake) err(op string) error {
	if f.failOn == op {
		return errors.New("boom")
	}
	return nil
}

func TestClampRate(t *testing.T) {
	cases := map[float64]float64{-1: MinRate, 0.1: MinRate, 0.5: 0.5, 2.5: 2.5, 8: 8, 12: MaxRate}
	for in, want := range cases {
		if got := ClampRate(in); got != want {
			t.Fatalf("ClampRate(%v)=%v want %v", in, got, want)
		}
	}
}

func TestToggleAllPausesWhenAnyPlaying(t *testing.T) {
	c := NewController()
	a, b := &fake{}, &fake{paused: true}
	c.Add(a)
	c.Add(b)
	paused, err := c.ToggleAll()
	if err != nil || !paused {
		t.Fatalf("ToggleAll=%v,%v want true,nil", paused, err)
	}
	if !a.paused || !b.paused {
		t.Fatal("not all paused")
	}
	paused, err = c.ToggleAll()
	if err != nil || paused {
		t.Fatalf("second ToggleAll=%v,%v want false,nil", paused, err)
	}
	if a.paused || b.paused {
		t.Fatal("not all playing")
	}
}

func TestSeekAllClampsPerPane(t *testing.T) {
	short := &fake{pos: 500, dur: 1000, rate: 1}
	long := &fake{pos: 5000, dur: 0, rate: 1} // unknown duration
	c := NewController()
	c.Add(short)
	c.Add(long)
	if err := c.SeekAll(-600); err != nil {
		t.Fatal(err)
	}
	if short.pos != 0 {
		t.Fatalf("short pos=%d want 0 (clamped)", short.pos)
	}
	if long.pos != 4400 {
		t.Fatalf("long pos=%d want 4400", long.pos)
	}
	if err := c.SeekAll(10_000); err != nil {
		t.Fatal(err)
	}
	if short.pos != 1000 {
		t.Fatalf("short pos=%d want duration 1000", short.pos)
	}
	if long.pos != 14400 {
		t.Fatalf("long pos=%d want 14400 (no duration clamp)", long.pos)
	}
}

func TestSetBaseAllAndNudgeClamp(t *testing.T) {
	a, b := &fake{rate: 1}, &fake{rate: 2}
	c := NewController()
	c.Add(a)
	c.Add(b)
	if err := c.SetBaseAll(3); err != nil {
		t.Fatal(err)
	}
	if a.rate != 3 || b.rate != 3 {
		t.Fatalf("base all=%v/%v want 3/3", a.rate, b.rate)
	}
	if err := c.NudgeAll(6); err != nil {
		t.Fatal(err)
	}
	if a.rate != MaxRate || b.rate != MaxRate {
		t.Fatalf("nudge up=%v/%v want %v", a.rate, b.rate, MaxRate)
	}
	if err := c.NudgeAll(-100); err != nil {
		t.Fatal(err)
	}
	if a.rate != MinRate || b.rate != MinRate {
		t.Fatalf("nudge down=%v/%v want %v", a.rate, b.rate, MinRate)
	}
}

func TestRemoveAtKeepsOrder(t *testing.T) {
	a, b, d := &fake{rate: 1}, &fake{rate: 2}, &fake{rate: 3}
	c := NewController()
	c.Add(a)
	c.Add(b)
	c.Add(d)
	c.RemoveAt(1)
	if c.Len() != 2 || c.At(0) != a || c.At(1) != d {
		t.Fatalf("remove kept wrong players: len=%d", c.Len())
	}
	c.RemoveAt(99)
	if c.Len() != 2 {
		t.Fatal("out-of-range remove changed length")
	}
}

func TestEmptyControllerIsPaused(t *testing.T) {
	c := NewController()
	if !c.Paused() {
		t.Fatal("empty controller should count as paused")
	}
	if paused, err := c.ToggleAll(); err != nil || paused {
		t.Fatalf("ToggleAll on empty=%v,%v want false,nil", paused, err)
	}
}

func TestEachReportsFirstError(t *testing.T) {
	a := &fake{failOn: "pause"}
	b := &fake{}
	c := NewController()
	c.Add(a)
	c.Add(b)
	if err := c.PauseAll(); err == nil {
		t.Fatal("want error")
	}
	if !b.paused {
		t.Fatal("failing player must not stop the rest")
	}
}
