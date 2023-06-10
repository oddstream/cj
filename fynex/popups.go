package fynex

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// func ShowListPopUp(canvas fyne.Canvas, title string, strs []string, okCallback func(string)) {
// 	var pu *widget.PopUp
// 	hdr := widget.NewLabel(title)
// 	sel := widget.NewSelect(strs, func(str string) {
// 		okCallback(str)
// 		pu.Hide()
// 	})
// 	cancel := widget.NewButton("Cancel", func() {
// 		pu.Hide()
// 	})
// 	content := container.New(layout.NewBorderLayout(hdr, cancel, nil, nil), hdr, cancel, sel)
// 	pu = widget.NewModalPopUp(content, canvas)
// 	pu.Show()
// }

func ShowListPopUp2(canvas fyne.Canvas, title string, strs []string, okCallback func(string)) {
	var pu *widget.PopUp
	hdr := widget.NewLabel(title)
	lbox := widget.NewList(
		func() int {
			return len(strs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(strs[id])
		},
	)
	lbox.OnSelected = func(id int) {
		okCallback(strs[id])
		pu.Hide()
	}
	cancel := widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	content := container.New(layout.NewBorderLayout(hdr, cancel, nil, nil), hdr, lbox, cancel)
	pu = widget.NewModalPopUp(content, canvas)
	pu.Resize(fyne.NewSize(200, 320))
	pu.Show()
}

func ShowListEntryPopUp(canvas fyne.Canvas, title string, strs []string, okCallback func(string)) {
	var pu *widget.PopUp
	hdr := widget.NewLabel(title)
	var currSel string
	sel := widget.NewSelect(strs, func(str string) {
		currSel = str
	})
	ent := widget.NewEntry()
	ent.PlaceHolder = "New"
	guts := container.New(layout.NewVBoxLayout(), sel, ent)
	ok := widget.NewButton("OK", func() {
		txt := ent.Text
		if txt == "" {
			txt = currSel
		}
		okCallback(txt)
		pu.Hide()
	})
	cancel := widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	bottom := container.New(layout.NewGridLayout(2), ok, cancel)
	content := container.New(layout.NewBorderLayout(hdr, bottom, nil, nil), hdr, bottom, guts)
	pu = widget.NewModalPopUp(content, canvas)
	pu.Show()
}

func ShowMarkdownPopup(parentCanvas fyne.Canvas, text string) {
	mkdn := widget.NewRichTextFromMarkdown(text)
	content := container.New(layout.NewBorderLayout(nil, nil, nil, nil), mkdn)
	pu := widget.NewPopUp(content, parentCanvas)
	pu.Show()
}

var popupMenu *widget.PopUpMenu

func ShowMenuPopup(parentCanvas fyne.Canvas, parent *widget.Toolbar) {
	// NewPopUpMenu creates a new, reusable popup menu. You can show it using ShowAtPosition
	if popupMenu == nil {
		menuItem1 := fyne.NewMenuItem("A", nil)
		menuItem2 := fyne.NewMenuItem("B", nil)
		menuItem3 := fyne.NewMenuItem("C", nil)
		menu := fyne.NewMenu("File", menuItem1, menuItem2, menuItem3)
		popupMenu = widget.NewPopUpMenu(menu, parentCanvas)
	}
	popupMenu.ShowAtPosition(parent.Position())
	// popupMenu.Show()
}
