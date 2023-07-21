package fynex

import (
	_ "embed"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// https://fonts.google.com/icons
// https://www.iconsdb.com/white-icons/
// https://www.iconsdb.com/white-icons/white-tag-icons.html
//
//go:embed "icons/tag-5-48.png"
var tagIconBytes []byte

/*
//go:embed "fonts/Hack-Regular.ttf"
var fontBytes []byte

var resourceFontTtf = &fyne.StaticResource{
	StaticName:    "Hack.ttf",
	StaticContent: fontBytes,
}
*/

type NoteTheme struct {
	colors map[fyne.ThemeColorName]color.RGBA
	sizes  map[fyne.ThemeSizeName]float32
	icons  map[fyne.ThemeIconName]fyne.Resource
}

var _ fyne.Theme = (*NoteTheme)(nil)

func NewNoteTheme() *NoteTheme {
	nt := &NoteTheme{}
	nt.colors = make(map[fyne.ThemeColorName]color.RGBA)
	// the Note color is now the color of the Note window background
	// and the "inputBackground" is always transparent to allow the background color to be seen
	nt.colors["inputBackground"] = color.RGBA{0, 0, 0, 0} // color.Transparent is 0,0,0,0
	// primary is the color used for caret and border of focussed widget
	// nt.colors["inputBorder"] = color.RGBA{255, 0, 0, 255}
	// nt.colors["selection"] = color.RGBA{255, 0, 0, 255}
	// fyne.CurrentApp().Settings().ThemeVariant()
	// VariantDark fyne.ThemeVariant = 0
	// VariantLight fyne.ThemeVariant = 1
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		nt.colors["primary"] = color.RGBA{255, 255, 255, 255}
	} else {
		nt.colors["primary"] = color.RGBA{0, 0, 0, 255}
	}
	// caret visibility https://github.com/fyne-io/fyne/issues/4063

	nt.icons = make(map[fyne.ThemeIconName]fyne.Resource)
	nt.icons["tag"] = &fyne.StaticResource{
		StaticName:    "tag.png",
		StaticContent: tagIconBytes,
	}
	return nt
}

func (nt *NoteTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// see https://github.com/fyne-io/fyne/blob/master/theme/theme.go
	// for color name and variant constants, eg
	// primary foreground placeholder button inputBorder inputBackground hover separator shadow scrollBar background
	// VariantDark fyne.ThemeVariant = 0
	// VariantLight fyne.ThemeVariant = 1
	if rgba, ok := nt.colors[name]; ok {
		return color.Color(rgba)
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (nt *NoteTheme) Font(s fyne.TextStyle) fyne.Resource {
	/*
		bundleFont("NotoSans-Regular.ttf", "regular", f)
		bundleFont("NotoSans-Bold.ttf", "bold", f)
		bundleFont("NotoSans-Italic.ttf", "italic", f)
		bundleFont("NotoSans-BoldItalic.ttf", "bolditalic", f)
		bundleFont("DejaVuSansMono-Powerline.ttf", "monospace", f)
	*/
	return theme.DefaultTextMonospaceFont()
	// return resourceFontTtf
	// return theme.DefaultTextFont()
}

func (nt *NoteTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if ico, ok := nt.icons[name]; ok {
		return ico
	}
	return theme.DefaultTheme().Icon(name)
}

func (nt *NoteTheme) Size(name fyne.ThemeSizeName) float32 {
	if sz, ok := nt.sizes[name]; ok {
		return sz
	}
	return theme.DefaultTheme().Size(name)
}
