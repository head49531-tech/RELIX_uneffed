package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type LabTheme struct {
	fallback fyne.Theme
}

func NewLabTheme() fyne.Theme {
	return &LabTheme{fallback: theme.DarkTheme()}
}

func (t *LabTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{14, 18, 22, 255}
	case theme.ColorNameButton:
		return color.NRGBA{26, 32, 40, 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{80, 90, 100, 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{20, 26, 32, 255}
	case theme.ColorNameError:
		return color.NRGBA{220, 90, 110, 255}
	case theme.ColorNameForeground:
		return color.NRGBA{220, 228, 236, 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{14, 18, 22, 255}
	case theme.ColorNameHover:
		return color.NRGBA{38, 46, 56, 255}
	case theme.ColorNameHyperlink:
		return color.NRGBA{64, 196, 180, 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{18, 24, 30, 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{50, 60, 72, 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{100, 110, 125, 255}
	case theme.ColorNamePressed:
		return color.NRGBA{52, 90, 84, 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{64, 196, 180, 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{50, 60, 72, 255}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{20, 26, 32, 120}
	case theme.ColorNameSelection:
		return color.NRGBA{64, 196, 180, 80}
	case theme.ColorNameSeparator:
		return color.NRGBA{40, 50, 60, 255}
	case theme.ColorNameShadow:
		return color.NRGBA{0, 0, 0, 120}
	case theme.ColorNameSuccess:
		return color.NRGBA{90, 210, 130, 255}
	case theme.ColorNameWarning:
		return color.NRGBA{240, 190, 70, 255}
	}
	return t.fallback.Color(name, variant)
}

func (t *LabTheme) Font(s fyne.TextStyle) fyne.Resource   { return t.fallback.Font(s) }
func (t *LabTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.fallback.Icon(n) }
func (t *LabTheme) Size(n fyne.ThemeSizeName) float32    { return t.fallback.Size(n) }
