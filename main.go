package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed icons/book-48.png
var book48IconBytes []byte // https://www.iconsdb.com/white-icons/book-icon.html

func main() {
	a := app.NewWithID("oddstream.goldnotebook")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "book-48.png",
		StaticContent: book48IconBytes,
	})
	w := a.NewWindow("Gold Notebook")
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}
