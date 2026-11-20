package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

// Icon is generated at startup instead of being committed as a binary, so the
// repository stays text-only.
var Icon = fyne.NewStaticResource("icon.png", buildIcon())

func buildIcon() []byte {
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{0x17, 0x18, 0x1a, 0xff}
	pane := color.RGBA{0x1e, 0x20, 0x23, 0xff}
	accent := color.RGBA{0x5b, 0x8c, 0xff, 0xff}

	fill := func(x0, y0, x1, y1 int, c color.RGBA) {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}
	fill(0, 0, size, size, bg)
	const m, gap = 28, 6
	fill(m, m, size-m, size-m, pane)
	half := (size - 2*m - gap) / 2
	cells := [][2]int{{m, m}, {m + half + gap, m}, {m, m + half + gap}, {m + half + gap, m + half + gap}}
	for i, c := range cells {
		col := pane
		if i == 0 {
			col = accent
		}
		fill(c[0], c[1], c[0]+half, c[1]+half, col)
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
