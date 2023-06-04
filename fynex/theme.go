package fynex

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed icons/book-128.png
var bookIconBytes []byte // https://www.iconsdb.com/white-icons/book-icon.html

type NoteTheme struct {
}

var _ fyne.Theme = (*NoteTheme)(nil)

func (nt *NoteTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (nt *NoteTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTextMonospaceFont()
	// return theme.TextMonospaceFont()
}

func (nt *NoteTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (nt *NoteTheme) Size(n fyne.ThemeSizeName) float32 {

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

func (nt *NoteTheme) BookIcon() fyne.Resource {
	return &fyne.StaticResource{
		StaticName:    "book.png",
		StaticContent: bookIconBytes,
	}
}
