//go:build cgo && mpv

// Package player: libmpv-backed implementation. Built with -tags "cgo mpv"
// (cgo is on by default; the mpv tag keeps the pure-Go tests independent of
// libmpv being installed).
package player

/*
#cgo pkg-config: mpv
#include <mpv/client.h>
#include <stdlib.h>
#include <string.h>

static int set_str(mpv_handle *h, const char *name, const char *val) {
    return mpv_set_property_string(h, name, val);
}
static int set_flag(mpv_handle *h, const char *name, int v) {
    return mpv_set_property(h, name, MPV_FORMAT_FLAG, &v);
}
static int set_double(mpv_handle *h, const char *name, double v) {
    return mpv_set_property(h, name, MPV_FORMAT_DOUBLE, &v);
}
static int get_double(mpv_handle *h, const char *name, double *out) {
    return mpv_get_property(h, name, MPV_FORMAT_DOUBLE, out);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// Engine is one libmpv instance. Each pane owns its own Engine so a slow file
// cannot stall the others.
type Engine struct {
	mu     sync.Mutex
	handle *C.mpv_handle
	path   string
	closed bool
}

// NewDefault returns a libmpv-backed Player.
func NewDefault() (Player, error) { return NewEngine() }

// NewEngine creates and initialises a libmpv instance without its own window.
// Rendering is done separately through the render API; for headless spikes the
// caller can set vo=null via SetOption before Open.
func NewEngine() (*Engine, error) {
	h := C.mpv_create()
	if h == nil {
		return nil, errors.New("player: mpv_create failed")
	}
	if rc := C.mpv_initialize(h); rc < 0 {
		C.mpv_terminate_destroy(h)
		return nil, fmt.Errorf("player: mpv_initialize failed: %d", int(rc))
	}
	return &Engine{handle: h}, nil
}

// SetOption sets a string option before Open (e.g. "vo" = "null").
func (e *Engine) SetOption(name, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	cn, cv := C.CString(name), C.CString(value)
	defer C.free(unsafe.Pointer(cn))
	defer C.free(unsafe.Pointer(cv))
	if rc := C.set_str(e.handle, cn, cv); rc < 0 {
		return fmt.Errorf("player: set %s=%s failed: %d", name, value, int(rc))
	}
	return nil
}

// Open loads a file and starts paused at position 0.
func (e *Engine) Open(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	cmd := fmt.Sprintf("loadfile %q", path)
	cc := C.CString(cmd)
	defer C.free(unsafe.Pointer(cc))
	if rc := C.mpv_command_string(e.handle, cc); rc < 0 {
		return fmt.Errorf("player: loadfile failed: %d", int(rc))
	}
	e.path = path
	return nil
}

// Close terminates the instance.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	C.mpv_terminate_destroy(e.handle)
	return nil
}

// Play starts playback.
func (e *Engine) Play() error { return e.setFlag("pause", false) }

// Pause pauses playback.
func (e *Engine) Pause() error { return e.setFlag("pause", true) }

// Paused reports the pause property.
func (e *Engine) Paused() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	var v C.int
	name := C.CString("pause")
	defer C.free(unsafe.Pointer(name))
	if rc := C.mpv_get_property(e.handle, name, C.MPV_FORMAT_FLAG, unsafe.Pointer(&v)); rc < 0 {
		return true
	}
	return v != 0
}

// Seek moves to an absolute position in milliseconds (exact seek).
func (e *Engine) Seek(ms int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	secs := float64(ms) / 1000.0
	args := []string{"seek", fmt.Sprintf("%.3f", secs), "absolute+exact"}
	cargs := make([]*C.char, len(args)+1)
	for i, a := range args {
		cargs[i] = C.CString(a)
		defer C.free(unsafe.Pointer(cargs[i]))
	}
	cargs[len(args)] = nil
	if rc := C.mpv_command(e.handle, (**C.char)(unsafe.Pointer(&cargs[0]))); rc < 0 {
		return fmt.Errorf("player: seek failed: %d", int(rc))
	}
	return nil
}

// Position returns the current position in milliseconds.
func (e *Engine) Position() int64 { return int64(e.getDouble("time-pos") * 1000) }

// Duration returns the file duration in milliseconds (0 when unknown).
func (e *Engine) Duration() int64 { return int64(e.getDouble("duration") * 1000) }

// SetRate sets playback speed (libmpv keeps audio pitch by default).
func (e *Engine) SetRate(rate float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	v := C.double(ClampRate(rate))
	name := C.CString("speed")
	defer C.free(unsafe.Pointer(name))
	if rc := C.set_double(e.handle, name, v); rc < 0 {
		return fmt.Errorf("player: set speed failed: %d", int(rc))
	}
	return nil
}

// Rate returns the current speed.
func (e *Engine) Rate() float64 { return e.getDouble("speed") }

// WaitEvent blocks for one mpv event (used by spikes to wait for file-loaded).
func (e *Engine) WaitEvent() (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ev := C.mpv_wait_event(e.handle, 0.1)
	return int(ev.event_id), nil
}

func (e *Engine) getDouble(name string) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	var v C.double
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	if rc := C.get_double(e.handle, cn, &v); rc < 0 {
		return 0
	}
	return float64(v)
}

func (e *Engine) setFlag(name string, on bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	v := 0
	if on {
		v = 1
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	if rc := C.set_flag(e.handle, cn, C.int(v)); rc < 0 {
		return fmt.Errorf("player: set %s failed: %d", name, int(rc))
	}
	return nil
}
