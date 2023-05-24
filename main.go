package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icons/book-48.png
var book48IconBytes []byte // https://www.iconsdb.com/white-icons/book-icon.html

type ui struct {
	current *note
	found   []*note

	calendar     *fyne.Container //*Calendar
	searchEntry  *widget.Entry
	searchButton *widget.Button
	foundList    *widget.List
	noteEntry    *widget.Entry
}

var theUI *ui = &ui{current: loadDatedNote(time.Now())}

func saveDirtyNote() {
	newText := theUI.noteEntry.Text
	if newText != theUI.current.text {
		theUI.current.text = newText
		theUI.current.save()
	}
}

func calendarTapped(t time.Time) {
	fmt.Println("calendar callback", t)
	saveDirtyNote()
	theUI.current = loadDatedNote(t)
	theUI.noteEntry.SetText(theUI.current.text)
}

func findButtonTapped() {
	query := theUI.searchEntry.Text
	if query == "" {
		return
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	theUI.found = []*note{}

	datedPath := path.Join(userHomeDir, NOTEBOOK_DIR, "dated")
	undatedPath := path.Join(userHomeDir, NOTEBOOK_DIR, "undated")
	results := Search(query, []string{datedPath, undatedPath})
	for _, fname := range results {
		theUI.found = append(theUI.found, load(fname))
	}
	theUI.foundList.Refresh()
}

func listSelected(id widget.ListItemID) {
	theUI.current = theUI.found[id]
	theUI.noteEntry.SetText(theUI.current.text)
}

func buildUI(u *ui) fyne.CanvasObject {
	u.calendar = container.New(layout.NewCenterLayout(), NewCalendar(time.Now(), calendarTapped))
	u.searchEntry = widget.NewEntry()
	u.searchButton = widget.NewButtonWithIcon("", theme.SearchIcon(), findButtonTapped)
	u.foundList = widget.NewList(
		func() int {
			return len(theUI.found)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// println("update widget.ListItemID", id)
			obj.(*widget.Label).SetText(theUI.found[id].title)
		},
	)
	u.foundList.OnSelected = listSelected

	// searchThings := container.New(layout.NewGridLayout(2), u.searchEntry, u.searchButton)
	sideTop := container.New(layout.NewVBoxLayout(), u.calendar, u.searchEntry, u.searchButton)
	sideBottom := container.New(layout.NewMaxLayout(), u.foundList)
	side := container.New(layout.NewBorderLayout(sideTop, nil, nil, nil), sideTop, sideBottom)

	u.noteEntry = widget.NewMultiLineEntry()
	u.noteEntry.SetText(theUI.current.text)

	return newAdaptiveSplit(side, u.noteEntry)
}

func main() {
	a := app.NewWithID("oddstream.goldnotebook")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "book-48.png",
		StaticContent: book48IconBytes,
	})
	w := a.NewWindow("Gold Notebook")
	w.SetContent(buildUI(theUI))
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	saveDirtyNote()
}
