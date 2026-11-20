package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Palette from the approved design (DESIGN.lock, variant A).
var (
	ColorBG     = color.NRGBA{R: 0x17, G: 0x18, B: 0x1a, A: 0xff}
	ColorChrome = color.NRGBA{R: 0x1a, G: 0x1c, B: 0x1f, A: 0xff}
	ColorPane   = color.NRGBA{R: 0x1e, G: 0x20, B: 0x23, A: 0xff}
	ColorLine   = color.NRGBA{R: 0x2a, G: 0x2d, B: 0x31, A: 0xff}
	ColorText   = color.NRGBA{R: 0xd7, G: 0xd9, B: 0xdd, A: 0xff}
	ColorMuted  = color.NRGBA{R: 0x8a, G: 0x8f, B: 0x98, A: 0xff}
	ColorAccent = color.NRGBA{R: 0x5b, G: 0x8c, B: 0xff, A: 0xff}
	ColorAccentWeak = color.NRGBA{R: 0x5b, G: 0x8c, B: 0xff, A: 0x28}
	ColorChipBG     = color.NRGBA{R: 0x0a, G: 0x0b, B: 0x0d, A: 0x8c}
	ColorHUD        = color.NRGBA{R: 0x0a, G: 0x0b, B: 0x0d, A: 0x9e}
)

// Theme overrides the handful of Fyne colours the spec fixes, falling back to
// the default dark theme for everything else.
type Theme struct{ base fyne.Theme }

// NewTheme returns the app theme.
func NewTheme() *Theme { return &Theme{base: theme.DarkTheme()} }

var _ fyne.Theme = (*Theme)(nil)

func (t *Theme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBG
	case theme.ColorNameForeground:
		return ColorText
	case theme.ColorNamePrimary:
		return ColorAccent
	case theme.ColorNameButton:
		return ColorChrome
	case theme.ColorNameDisabledButton:
		return ColorPane
	case theme.ColorNameInputBackground:
		return ColorPane
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return ColorChrome
	case theme.ColorNameInputBorder:
		return ColorLine
	case theme.ColorNameSeparator:
		return ColorLine
	case theme.ColorNamePlaceHolder:
		return ColorMuted
	case theme.ColorNameDisabled:
		return ColorMuted
	case theme.ColorNameScrollBar:
		return ColorLine
	case theme.ColorNameHover:
		return ColorAccentWeak
	case theme.ColorNameSelection:
		return ColorAccentWeak
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 2
	case theme.SizeNameText:
		return 12
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 3
	case theme.SizeNameScrollBar, theme.SizeNameScrollBarSmall:
		return 4
	}
	return t.base.Size(name)
}
