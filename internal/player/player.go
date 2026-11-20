// Package player defines the playback abstraction and the global-control
// semantics (pause all, seek all, rate overlay) independent of the concrete
// engine. The libmpv implementation lives in engine_mpv.go behind a build tag
// so this package stays testable without cgo.
package player

import (
	"fmt"
	"image"
)

// Rate limits from the requirements: every pane can run at 0.5x..8x.
const (
	MinRate = 0.5
	MaxRate = 8.0
)

// ClampRate keeps a rate inside [MinRate, MaxRate].
func ClampRate(r float64) float64 {
	if r < MinRate {
		return MinRate
	}
	if r > MaxRate {
		return MaxRate
	}
	return r
}

// VideoRenderer is implemented by engines that can hand out decoded frames as
// images (libmpv software render API). The UI paints them into a canvas.Image.
type VideoRenderer interface {
	// RenderFrame draws the current frame at w×h and returns it. The second
	// return value reports whether a new frame was produced.
	RenderFrame(w, h int) (image.Image, bool)
}

// Player is one video pane's playback control.
type Player interface {
	Open(path string) error
	Close() error
	Play() error
	Pause() error
	Paused() bool
	// Seek moves to an absolute position; implementations clamp to the file.
	SeekTo(ms int64) error
	Position() int64 // ms
	Duration() int64 // ms, 0 when unknown
	SetRate(rate float64) error
	Rate() float64
}

// Controller applies the global controls to every pane.
//
// Semantics (approved in UI-SPEC): global seek is applied independently to
// each pane and clamped to that pane's duration; there is no cross-pane clock
// synchronisation.
type Controller struct {
	players []Player
}

// NewController returns an empty controller.
func NewController() *Controller { return &Controller{} }

// Add appends a player and returns its index.
func (c *Controller) Add(p Player) int {
	c.players = append(c.players, p)
	return len(c.players) - 1
}

// RemoveAt drops a player, keeping the relative order of the rest.
func (c *Controller) RemoveAt(i int) {
	if i < 0 || i >= len(c.players) {
		return
	}
	c.players = append(c.players[:i], c.players[i+1:]...)
}

// Len reports the number of panes.
func (c *Controller) Len() int { return len(c.players) }

// At returns the player at i, or nil.
func (c *Controller) At(i int) Player {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// Paused reports whether every pane is paused (an empty set counts as paused).
func (c *Controller) Paused() bool {
	for _, p := range c.players {
		if !p.Paused() {
			return false
		}
	}
	return true
}

// ToggleAll pauses every pane when any pane is playing, otherwise plays all.
// It returns the resulting paused state.
func (c *Controller) ToggleAll() (bool, error) {
	if !c.Paused() {
		return true, c.PauseAll()
	}
	return false, c.PlayAll()
}

// PauseAll pauses every pane, reporting the first error but attempting all.
func (c *Controller) PauseAll() error { return c.each(func(p Player) error { return p.Pause() }) }

// PlayAll plays every pane.
func (c *Controller) PlayAll() error { return c.each(func(p Player) error { return p.Play() }) }

// SeekAll moves every pane by deltaMs, clamped to each pane's own duration.
func (c *Controller) SeekAll(deltaMs int64) error {
	return c.each(func(p Player) error {
		target := p.Position() + deltaMs
		if target < 0 {
			target = 0
		}
		if d := p.Duration(); d > 0 && target > d {
			target = d
		}
		return p.SeekTo(target)
	})
}

// SetBaseAll sets the same rate on every pane (the "global speed" dropdown).
func (c *Controller) SetBaseAll(rate float64) error {
	r := ClampRate(rate)
	return c.each(func(p Player) error { return p.SetRate(r) })
}

// NudgeAll adds delta to every pane's current rate (the "overlay" stepper).
func (c *Controller) NudgeAll(delta float64) error {
	return c.each(func(p Player) error { return p.SetRate(ClampRate(p.Rate() + delta)) })
}

func (c *Controller) each(fn func(Player) error) error {
	var first error
	for _, p := range c.players {
		if err := fn(p); err != nil && first == nil {
			first = fmt.Errorf("player: %w", err)
		}
	}
	return first
}
