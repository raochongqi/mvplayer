//go:build cgo && mpv

package player

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenderFrameProducesPicture(t *testing.T) {
	file := os.Getenv("MVPLAYER_SAMPLE")
	if file == "" {
		t.Skip("set MVPLAYER_SAMPLE to a real .mkv to run the render test")
	}
	e, err := NewEngine()
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if err := e.SetOption("ao", "null"); err != nil {
		t.Fatal(err)
	}
	if err := e.Open(file); err != nil {
		t.Fatal(err)
	}
	if err := e.Play(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1200 * time.Millisecond)

	img, ok := e.RenderFrame(320, 180)
	if !ok {
		t.Fatal("no frame rendered")
	}
	b := img.Bounds()
	if b.Dx() != 320 || b.Dy() != 180 {
		t.Fatalf("frame size %v", b)
	}
	// A black frame has zero variance; the generated clip is a moving gradient.
	var sum, sum2, n int
	for y := 0; y < b.Dy(); y += 7 {
		for x := 0; x < b.Dx(); x += 7 {
			r, g, bl, a := img.At(x, y).RGBA()
			if a>>8 != 255 {
				t.Fatalf("alpha=%d at %d,%d", a>>8, x, y)
			}
			v := float64(r>>8) + float64(g>>8) + float64(bl>>8)
			sum += int(v)
			sum2 += int(v * v)
			n++
		}
	}
	mean := float64(sum) / float64(n)
	variance := float64(sum2)/float64(n) - mean*mean
	if variance < 50 {
		t.Fatalf("frame looks blank, variance=%.1f mean=%.1f", variance, mean)
	}
	f, _ := os.Create(filepath.Join(t.TempDir(), "frame-check.png"))
	png.Encode(f, img)
	f.Close()
	t.Logf("frame ok variance=%.1f mean=%.1f", variance, mean)
}
