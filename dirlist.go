package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type DirList struct {
	widget.List
}

func (dl *DirList) MinSize() fyne.Size {
	return fyne.Size{Width: 32, Height: 64}
}

func NewDirList(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.ListItemID, fyne.CanvasObject)) *DirList {
	dl := &DirList{}
	dl.ExtendBaseWidget(dl)
	return dl
	// return widget.NewList(length, createItem, updateItem)
}
