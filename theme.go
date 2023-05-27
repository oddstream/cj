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
	// return theme.TextMonospaceFont()
}

func (m *noteTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m *noteTheme) Size(n fyne.ThemeSizeName) float32 {

	if n == theme.SizeNameText { // default SizeNameText is 12
		// fmt.Println(theme.DefaultTheme().Size(n))
		return theme.DefaultTheme().Size(n) + 2
	}

	// if n == theme.SizeNamePadding {	// default Padding is 6
	// fmt.Println(theme.DefaultTheme().Size(n))
	// }

	// if n == theme.SizeNameInnerPadding {	// default InnerPadding is 8
	// fmt.Println(theme.DefaultTheme().Size(n))
	// }

	if n == theme.SizeNameLineSpacing { // default LineSpacing is 4
		return theme.DefaultTheme().Size(n) + 2
	}
	return theme.DefaultTheme().Size(n)
}
