package fynex

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowListPopUp(canvas fyne.Canvas, title string, strs []string, okCallback func(string)) {
	var pu *widget.PopUp
	// var currSel int
	// lbox := widget.NewList(
	// 	func() int {
	// 		return len(testStrings)
	// 	},
	// 	func() fyne.CanvasObject {
	// 		return widget.NewLabel("")
	// 	},
	// 	func(id widget.ListItemID, obj fyne.CanvasObject) {
	// 		obj.(*widget.Label).SetText(testStrings[id])
	// 	},
	// )
	// lbox.Select(currSel)
	// lbox.OnSelected = func(id int) {
	// 	currSel = id
	// }
	hdr := widget.NewLabel(title)
	var currSel string
	sel := widget.NewSelect(strs, func(str string) {
		currSel = str
	})
	ok := widget.NewButton("OK", func() {
		// okCallback(testStrings[currSel])
		okCallback(currSel)
		pu.Hide()
	})
	cancel := widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	bottom := container.New(layout.NewGridLayout(2), ok, cancel)
	content := container.New(layout.NewBorderLayout(hdr, bottom, nil, nil), hdr, bottom, sel)
	pu = widget.NewModalPopUp(content, canvas)
	pu.Show()
}

func ShowListEntryPopUp() string {
	return ""
}
