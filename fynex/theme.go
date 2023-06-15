package fynex

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed icons/book-128.png
var bookIconBytes []byte // https://www.iconsdb.com/white-icons/book-icon.html

//go:embed icons/today-128.png
var todayIconBytes []byte // https://www.iconsdb.com/white-icons/today-icon.html

//go:embed "fonts/Hack-Regular.ttf"
var fontBytes []byte

var resourceFontTtf = &fyne.StaticResource{
	StaticName:    "Hack.ttf",
	StaticContent: fontBytes,
}

type NoteTheme struct {
	FontSize float32
	IconName string
}

var _ fyne.Theme = (*NoteTheme)(nil)

func (nt *NoteTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (nt *NoteTheme) Font(s fyne.TextStyle) fyne.Resource {
	// return theme.DefaultTextMonospaceFont()
	// return theme.TextMonospaceFont()
	return resourceFontTtf
}

func (nt *NoteTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (nt *NoteTheme) Size(n fyne.ThemeSizeName) float32 {

	if n == theme.SizeNameText { // default SizeNameText is 12 or 13
		// fmt.Println(theme.DefaultTheme().Size(n))
		// return theme.DefaultTheme().Size(n)
		if nt.FontSize == 0.0 {
			return theme.DefaultTheme().Size(n)
		} else {
			return nt.FontSize
		}
	}
	/*
		// tweaking these does NOT stop the top of some fonts being clipped
		// visible with Comic Mono

		if n == theme.SizeNamePadding { // default Padding is 6
			// fmt.Println(theme.DefaultTheme().Size(n))
			return theme.DefaultTheme().Size(n) + 4
		}

		if n == theme.SizeNameInnerPadding { // default InnerPadding is 8
			// fmt.Println(theme.DefaultTheme().Size(n))
			return theme.DefaultTheme().Size(n) + 4
		}

		if n == theme.SizeNameLineSpacing { // default LineSpacing is 4
			// fmt.Println(theme.DefaultTheme().Size(n))
			return theme.DefaultTheme().Size(n) + 4
		}
	*/

	return theme.DefaultTheme().Size(n)
}

func (nt *NoteTheme) BookIcon() fyne.Resource {
	if nt.IconName == "book" {
		return &fyne.StaticResource{
			StaticName:    "book.png",
			StaticContent: bookIconBytes,
		}
	} else if nt.IconName == "today" {
		return &fyne.StaticResource{
			StaticName:    "today.png",
			StaticContent: todayIconBytes,
		}
	} else {
		return theme.DefaultTheme().Icon(theme.IconNameComputer)
	}
}
