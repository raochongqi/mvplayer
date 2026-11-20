//go:build !mpv

package player

import (
	"sync"
	"time"
)

// NewDefault returns the libmpv engine when built with `-tags mpv`, otherwise
// an in-memory stub so the UI can be built and tested without libmpv.
func NewDefault() (Player, error) { return &stub{}, nil }

type stub struct {
	mu      sync.Mutex
	path    string
	paused  bool
	rate    float64
	pos     int64
	started time.Time
	base    int64
}

func (s *stub) Open(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.path, s.pos, s.base, s.paused, s.rate = path, 0, 0, false, 1
	s.started = time.Now()
	return nil
}
func (s *stub) Close() error { return nil }
func (s *stub) Play() error  { s.mu.Lock(); defer s.mu.Unlock(); s.paused = false; s.started = time.Now(); return nil }
func (s *stub) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.paused {
		s.pos = s.current()
		s.base = s.pos
	}
	s.paused = true
	return nil
}
func (s *stub) Paused() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.paused }
func (s *stub) current() int64 {
	if s.paused {
		return s.pos
	}
	return s.base + int64(time.Since(s.started).Seconds()*1000*s.rate)
}
func (s *stub) Seek(ms int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ms < 0 {
		ms = 0
	}
	s.pos, s.base, s.started = ms, ms, time.Now()
	return nil
}
func (s *stub) Position() int64 { s.mu.Lock(); defer s.mu.Unlock(); return s.current() }
func (s *stub) Duration() int64 { return 600_000 }
func (s *stub) SetRate(r float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pos, s.base, s.started = s.current(), s.current(), time.Now()
	s.rate = ClampRate(r)
	return nil
}
func (s *stub) Rate() float64 { s.mu.Lock(); defer s.mu.Unlock(); return s.rate }
