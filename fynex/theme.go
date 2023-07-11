package fynex

import (
	// _ "embed"
	"encoding/json"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

/*
//go:embed "fonts/Hack-Regular.ttf"
var fontBytes []byte

var resourceFontTtf = &fyne.StaticResource{
	StaticName:    "Hack.ttf",
	StaticContent: fontBytes,
}
*/

// BasicColors is a string-indexed map of the basic colors from HTML 4.01
var BasicColors = map[string]color.RGBA{
	"White":  {R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	"Silver": {R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff},
	"Gray":   {R: 0x80, G: 0x80, B: 0x80, A: 0xff},
	"Black":  {R: 0x00, G: 0x00, B: 0x00, A: 0xff},
	"Red":    {R: 0xff, G: 0, B: 0, A: 0xff},
	"Maroon": {R: 0x80, G: 0, B: 0, A: 0xff},
	"Yellow": {R: 0xff, G: 0xff, B: 0, A: 0xff},
	"Olive":  {R: 0x80, G: 0x80, B: 0, A: 0xff},
	"Lime":   {R: 0, G: 0xff, B: 0, A: 0xff},
	"Green":  {R: 0, G: 0x80, B: 0, A: 0xff},
	"Aqua":   {R: 0, G: 0xff, B: 0xff, A: 0xff},
	"Teal":   {R: 0, G: 0x80, B: 0x80, A: 0xff},
	"Blue":   {R: 0, G: 0, B: 0xff, A: 0xff},
	"Navy":   {R: 0, G: 0, B: 0x80, A: 0xff},
	"Fushia": {R: 0xff, G: 0, B: 0xff, A: 0xff},
	"Purple": {R: 0x80, G: 0, B: 0x80, A: 0xff},
}

// ExtendedColors see en.wikipedia.org/wiki/Web_colors
var ExtendedColors = map[string]color.RGBA{
	// Red
	"DarkRed":     {R: 0x8b, G: 0x00, B: 0x00, A: 0xff},
	"Red":         {R: 0xff, G: 0x00, B: 0x00, A: 0xff},
	"Firebrick":   {R: 0xb2, G: 0x22, B: 0x22, A: 0xff},
	"Crimson":     {R: 0xdc, G: 0x14, B: 0x3c, A: 0xff},
	"IndianRed":   {R: 0xcd, G: 0x5c, B: 0x5c, A: 0xff},
	"LightCoral":  {R: 0xf0, G: 0x80, B: 0x80, A: 0xff},
	"Salmon":      {R: 0xfa, G: 0x80, B: 0x72, A: 0xff},
	"DarkSalmon":  {R: 0xe9, G: 0x96, B: 0x7a, A: 0xff},
	"LightSalmon": {R: 0xff, G: 0xa0, B: 0x7a, A: 0xff},
	// Orange
	"OrangeRed":  {R: 0xff, G: 0x45, B: 0x00, A: 0xff},
	"Tomato":     {R: 0xff, G: 0x63, B: 0x47, A: 0xff},
	"DarkOrange": {R: 0xff, G: 0x8c, B: 0x00, A: 0xff},
	"Coral":      {R: 0xff, G: 0x7f, B: 0x50, A: 0xff},
	"Orange":     {R: 0xff, G: 0xA5, B: 0x00, A: 0xff},
	// Pink
	"MediumVioletRed": {R: 0xc7, G: 0x15, B: 0x85, A: 0xff},
	"DeepPink":        {R: 0xff, G: 0x14, B: 0x93, A: 0xff},
	"PaleVioletRed":   {R: 0xdb, G: 0x70, B: 0x93, A: 0xff},
	"HotPink":         {R: 0xff, G: 0x69, B: 0xb4, A: 0xff},
	"LightPink":       {R: 0xff, G: 0xb6, B: 0xc1, A: 0xff},
	"Pink":            {R: 0xff, G: 0xc0, B: 0xcb, A: 0xff},
	// Blue
	"Navy":           {R: 0, G: 0, B: 0x80, A: 0xff},
	"DarkBlue":       {R: 0, G: 0, B: 0x8b, A: 0xff},
	"MediumBlue":     {R: 0, G: 0, B: 0xcd, A: 0xff},
	"Blue":           {R: 0, G: 0, B: 0xff, A: 0xff},
	"MidnightBlue":   {R: 0x19, G: 0x19, B: 0x70, A: 0xff},
	"RoyalBlue":      {R: 0x41, G: 0x69, B: 0xe1, A: 0xff},
	"SteelBlue":      {R: 0x46, G: 0x82, B: 0xb4, A: 0xff},
	"DodgerBlue":     {R: 0x1e, G: 0x90, B: 0xff, A: 0xff},
	"DeepSkyBlue":    {R: 0x0, G: 0xbf, B: 0xff, A: 0xff},
	"CornflowerBlue": {R: 0x64, G: 0x95, B: 0xed, A: 0xff},
	"SkyBlue":        {R: 0x87, G: 0xce, B: 0xeb, A: 0xff},
	"LightSkyBlue":   {R: 0x87, G: 0xce, B: 0xfa, A: 0xff},
	"LightSteelBlue": {R: 0xb0, G: 0xc4, B: 0xde, A: 0xff},
	"LightBlue":      {R: 0xad, G: 0xd8, B: 0xe6, A: 0xff},
	"PowderBlue":     {R: 0xb0, G: 0xe0, B: 0xe6, A: 0xff},
	// Yellow
	"DarkKhaki":            {R: 0xbd, G: 0xb7, B: 0x6b, A: 0xff},
	"Gold":                 {R: 0xff, G: 0xd7, B: 0, A: 0xff},
	"Khaki":                {R: 0xf0, G: 0xe6, B: 0x8c, A: 0xff},
	"PeachPuff":            {R: 0xff, G: 0xda, B: 0xb9, A: 0xff},
	"Yellow":               {R: 0xff, G: 0xff, B: 0, A: 0xff},
	"PaleGoldenrod":        {R: 0xee, G: 0xe8, B: 0xaa, A: 0xff},
	"Moccasin":             {R: 0xff, G: 0xe4, B: 0xb5, A: 0xff},
	"PapayaWhip":           {R: 0xff, G: 0xef, B: 0xd5, A: 0xff},
	"LightGoldenrodYellow": {R: 0xfa, G: 0xfa, B: 0xd2, A: 0xff},
	"LemonChiffron":        {R: 0xff, G: 0xfa, B: 0xcd, A: 0xff},
	"LightYellow":          {R: 0xff, G: 0xff, B: 0xe0, A: 0xff},
	// Green
	"Green":       {R: 0, G: 0x80, B: 0, A: 0xff},
	"LightGreen":  {R: 0x90, G: 0xEE, B: 0x90, A: 0xff},
	"LimeGreen":   {R: 0x32, G: 0xCD, B: 0x32, A: 0xff},
	"DarkGreen":   {R: 0, G: 0x64, B: 0, A: 0xff},
	"BaizeGreen":  {R: 0, G: 0x50, B: 0, A: 0xff}, // okay, I added this one
	"ForestGreen": {R: 0x22, G: 0x8B, B: 0x22, A: 0xff},
	// Brown
	"Maroon":      {R: 0x80, G: 0x00, B: 0x00, A: 0xff},
	"Brown":       {R: 0xA5, G: 0x2A, B: 0x2A, A: 0xff},
	"SaddleBrown": {R: 0x8B, G: 0x45, B: 0x13, A: 0xff},
	"Sienna":      {R: 0xA0, G: 0x25, B: 0x2D, A: 0xff},
	"Chocolate":   {R: 0xD2, G: 0x69, B: 0x1E, A: 0xff},
	"Peru":        {R: 0xCD, G: 0x85, B: 0x3F, A: 0xff},
	"RosyBrown":   {R: 0xbc, G: 0x8f, B: 0x8f, A: 0xff},
	// TODO
	"Tan": {R: 0xD2, G: 0xB4, B: 0x8C, A: 0xff},
	// TODO
	// Purple
	"Indigo": {R: 0x4b, G: 0, B: 0x82, A: 0xff},
	"Purple": {R: 0x80, G: 0, B: 0x80, A: 0xff},
	// TODO
	// Gray
	"Black":          {R: 0, G: 0, B: 0, A: 0xff},
	"DarkSlateGray":  {R: 0x2f, G: 0x4f, B: 0x4f, A: 0xff},
	"DimGray":        {R: 0x69, G: 0x69, B: 0x69, A: 0xff},
	"SlateGray":      {R: 0x70, G: 0x80, B: 0x90, A: 0xff},
	"Gray":           {R: 0x80, G: 0x80, B: 0x80, A: 0xff},
	"LightSlateGray": {R: 0x77, G: 0x88, B: 0x99, A: 0xff},
	"DarkGray":       {R: 0xa9, G: 0xa9, B: 0xa9, A: 0xff},
	"Silver":         {R: 0xC0, G: 0xC0, B: 0xC0, A: 0xff},
	"LightGray":      {R: 0xD3, G: 0xD3, B: 0xD3, A: 0xff},
	"Gainsboro":      {R: 0xdc, G: 0xdc, B: 0xdc, A: 0xff},
	// White
	"WhiteSmoke":  {R: 0xF5, G: 0xF5, B: 0xF5, A: 0xff},
	"Honeydew":    {R: 0xF0, G: 0xFF, B: 0xF0, A: 0xff},
	"FloralWhite": {R: 0xFF, G: 0xFA, B: 0xF0, A: 0xff},
	"Azure":       {R: 0xF0, G: 0xFF, B: 0xFF, A: 0xff},
	"MintCream":   {R: 0xF5, G: 0xFF, B: 0xFA, A: 0xff},
	"Snow":        {R: 0xFA, G: 0xFA, B: 0xFA, A: 0xff},
	"Ivory":       {R: 0xFF, G: 0xFF, B: 0xF0, A: 0xff},
	"White":       {R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
	// TODO complete this map with all extended colors
} // golang gotcha no newline after last literal, must be comma or closing brace

type NoteTheme struct {
	Colors map[fyne.ThemeColorName]string
	Sizes  map[fyne.ThemeSizeName]float32
}

var _ fyne.Theme = (*NoteTheme)(nil)

func NewNoteTheme(pathname string) *NoteTheme {
	nt := &NoteTheme{}

	var jsonBytes []byte
	var err error
	if jsonBytes, err = os.ReadFile(pathname); err != nil {
		// log.Println(err)
	} else {
		// golang gotcha the len of the jsonBytes buffer must be the number of bytes actually read
		if err = json.Unmarshal(jsonBytes, nt); err != nil {
			log.Println(err)
		}
	}
	// a nil map can be read from, but it can't be written to
	// would need to create these maps if we ever dynamically add values to it
	// if nt.Colors == nil {
	// 	nt.Colors = make(map[fyne.ThemeColorName]string)
	// }
	// if nt.Sizes == nil {
	// 	nt.Sizes = make(map[fyne.ThemeSizeName]float32)
	// }

	return nt
}

func (nt *NoteTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// see https://github.com/fyne-io/fyne/blob/master/theme/theme.go
	// for color name and variant constants, eg
	// primary foreground placeholder button inputBorder inputBackground hover separator shadow scrollBar background
	// VariantDark fyne.ThemeVariant = 0
	// VariantLight fyne.ThemeVariant = 1
	if colName, ok := nt.Colors[name]; ok {
		if col, ok := BasicColors[colName]; ok {
			return col
		}
		if col, ok := ExtendedColors[colName]; ok {
			return col
		} else {
			log.Println("No Basic or Extended Color called", colName)
		}
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

func (nt *NoteTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (nt *NoteTheme) Size(name fyne.ThemeSizeName) float32 {
	if sz, ok := nt.Sizes[name]; ok {
		return sz
	}
	return theme.DefaultTheme().Size(name)
}
