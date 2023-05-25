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
	"fyne.io/fyne/v2/driver/desktop"
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

var theUI *ui = &ui{current: load(time.Now())}

func saveDirtyNote() {
	newText := theUI.noteEntry.Text
	if newText != theUI.current.text {
		theUI.current.text = newText
		theUI.current.save()
	}
}

func calendarTapped(t time.Time) {
	// fmt.Println("calendar callback", t)
	saveDirtyNote()
	theUI.current = load(t)
	theUI.noteEntry.SetText(theUI.current.text)

	theUI.foundList.UnselectAll()
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

	results := Search(query, []string{path.Join(userHomeDir, NOTEBOOK_DIR)})
	for _, fname := range results {
		fmt.Println("found", fname)
		date := parseDateFromFname(fname)
		theUI.found = append(theUI.found, load(date))
	}
	theUI.foundList.UnselectAll()
	theUI.foundList.Refresh()
}

func listSelected(id widget.ListItemID) {
	// log.Printf("list item %d selected", id)
	saveDirtyNote()
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
			obj.(*widget.Label).SetText(theUI.found[id].title())
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
	a.Settings().SetTheme(&noteTheme{})
	w := a.NewWindow("Gold Notebook")
	// shortcuts get swallowed if focus is in the note multiline entry widget
	ctrlF := &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		w.Canvas().Focus(theUI.searchEntry)
	})
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		saveDirtyNote()
	})
	w.SetContent(buildUI(theUI))
	w.Resize(fyne.NewSize(1024, 640))
	w.CenterOnScreen()
	w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	saveDirtyNote()
}
