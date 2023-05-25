package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type noteTheme struct {
}

func (t *noteTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (m *noteTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTextMonospaceFont()
}

func (m *noteTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m *noteTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return theme.DefaultTheme().Size(n) + 1
	}
	return theme.DefaultTheme().Size(n)
}
