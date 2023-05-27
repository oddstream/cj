package main

import (
	"fyne.io/fyne/v2"
)

type DirListLayout struct {
}

func (dll *DirListLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	// childSize := objects[0].MinSize()
	// return fyne.Size{Width: childSize.Width, Height: childSize.Height * 3}
	return fyne.Size{Width: 32, Height: 96}
}

func (dll *DirListLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	// pos := fyne.NewPos(0, 0)
	for _, o := range objects {
		size := o.MinSize()
		o.Resize(size)
		// o.Move(pos)
	}
}
