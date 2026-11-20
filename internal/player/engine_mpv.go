//go:build cgo && mpv

// Package player: libmpv-backed implementation. Built with -tags "cgo mpv"
// (cgo is on by default; the mpv tag keeps the pure-Go tests independent of
// libmpv being installed).
package player

/*
#cgo pkg-config: mpv
#include <mpv/client.h>
#include <mpv/render.h>
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

static int create_sw_render(mpv_handle *h, mpv_render_context **ctx) {
    char *api = "sw";
    mpv_render_param params[2];
    params[0].type = MPV_RENDER_PARAM_API_TYPE;
    params[0].data = api;
    params[1].type = MPV_RENDER_PARAM_INVALID;
    params[1].data = NULL;
    return mpv_render_context_create(ctx, h, params);
}

static int render_sw(mpv_render_context *ctx, int w, int h, void *buf, size_t stride) {
    int size[2];
    char *fmt = "rgb0";
    mpv_render_param params[5];
    size[0] = w;
    size[1] = h;
    params[0].type = MPV_RENDER_PARAM_SW_SIZE;
    params[0].data = size;
    params[1].type = MPV_RENDER_PARAM_SW_FORMAT;
    params[1].data = fmt;
    params[2].type = MPV_RENDER_PARAM_SW_STRIDE;
    params[2].data = &stride;
    params[3].type = MPV_RENDER_PARAM_SW_POINTER;
    params[3].data = buf;
    params[4].type = MPV_RENDER_PARAM_INVALID;
    params[4].data = NULL;
    return mpv_render_context_render(ctx, params);
}

static void free_render(mpv_render_context *ctx) { mpv_render_context_free(ctx); }
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
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

	rctx    *C.mpv_render_context
	bufs    [2][]byte
	bufIdx  int
	lastRGBA *image.RGBA
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
	// The render API needs the libmpv video output; keep audio pitch stable at
	// any speed (requirement: 0.5x–8x per pane).
	for _, kv := range [][2]string{
		{"vo", "libmpv"},
		{"audio-pitch-correction", "yes"},
		{"hwdec", "auto-safe"},
	} {
		k, v := C.CString(kv[0]), C.CString(kv[1])
		C.mpv_set_option_string(h, k, v)
		C.free(unsafe.Pointer(k))
		C.free(unsafe.Pointer(v))
	}
	if rc := C.mpv_initialize(h); rc < 0 {
		C.mpv_terminate_destroy(h)
		return nil, fmt.Errorf("player: mpv_initialize failed: %d", int(rc))
	}
	e := &Engine{handle: h}
	if rc := C.create_sw_render(h, &e.rctx); rc < 0 {
		C.mpv_terminate_destroy(h)
		return nil, fmt.Errorf("player: sw render context failed: %d", int(rc))
	}
	return e, nil
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
	if e.rctx != nil {
		C.free_render(e.rctx)
		e.rctx = nil
	}
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

// RenderFrame draws the current video frame at w×h into an RGBA image.
// The software renderer writes rgb0 (alpha 0); the alpha byte is filled with
// 255 so canvas.Image composites it.
func (e *Engine) RenderFrame(w, h int) (image.Image, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.rctx == nil || e.closed || w <= 0 || h <= 0 {
		return nil, false
	}
	n := w * h * 4
	idx := e.bufIdx
	if cap(e.bufs[idx]) < n {
		e.bufs[idx] = make([]byte, n)
	}
	buf := e.bufs[idx][:n]
	stride := C.size_t(w * 4)
	if C.render_sw(e.rctx, C.int(w), C.int(h), unsafe.Pointer(&buf[0]), stride) < 0 {
		return nil, false
	}
	for i := 3; i < n; i += 4 {
		buf[i] = 255
	}
	img := &image.RGBA{Pix: buf, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
	e.lastRGBA = img
	e.bufIdx = 1 - idx
	return img, true
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
