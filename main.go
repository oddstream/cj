package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	// popUp        *widget.PopUp
}

var (
	theUI          *ui
	theUserHomeDir string
	theDataDir     string
	theBookDir     string
	debugMode      bool
)

func (u *ui) saveDirtyNote() {
	newText := u.noteEntry.Text
	if newText != u.current.text {
		u.current.text = newText
		u.current.save()
	}
}

func (u *ui) setCurrent(n *note) {
	theUI.current = n
	theUI.noteEntry.SetText(theUI.current.text)
	theUI.calendar.Objects[0] = NewCalendar(theUI.current.date, calendarTapped, calendarIsDateImportant)
}

func calendarTapped(t time.Time) {
	// fmt.Println("calendar callback", t)
	theUI.saveDirtyNote()
	theUI.setCurrent(load(t))
	theUI.foundList.UnselectAll()
}

func calendarIsDateImportant(t time.Time) bool {
	return t.Year() == theUI.current.date.Year() &&
		t.Month() == theUI.current.date.Month() &&
		t.Day() == theUI.current.date.Day()
}

func find(query string) {
	if query == "" {
		return
	}
	// query = strings.ToLower(query)

	theUI.found = []*note{}

	results := Search(query, []string{path.Join(theUserHomeDir, theDataDir, theBookDir)})
	for _, fname := range results {
		if debugMode {
			fmt.Println("found", fname)
		}
		date := parseDateFromFname(fname)
		theUI.found = append(theUI.found, load(date))
	}
	theUI.foundList.UnselectAll()
	theUI.foundList.Refresh()
}

func findButtonTapped() {
	find(theUI.searchEntry.Text)
}

func listSelected(id widget.ListItemID) {
	// log.Printf("list item %d selected", id)
	theUI.saveDirtyNote()
	theUI.setCurrent(theUI.found[id])
}

func buildUI(u *ui) fyne.CanvasObject {
	u.calendar = container.New(layout.NewCenterLayout(), NewCalendar(theUI.current.date, calendarTapped, calendarIsDateImportant))
	u.searchEntry = widget.NewEntry()
	u.searchEntry.OnSubmitted = func(str string) {
		find(str)
	}
	u.searchEntry.TextStyle = fyne.TextStyle{Monospace: true}
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

	searchThings := container.New(layout.NewFormLayout(), u.searchButton, u.searchEntry)
	sideTop := container.New(layout.NewVBoxLayout(), u.calendar, searchThings)
	// sideTop := container.New(layout.NewVBoxLayout(), u.calendar, u.searchEntry, u.searchButton)
	sideBottom := container.New(layout.NewMaxLayout(), u.foundList)
	side := container.New(layout.NewBorderLayout(sideTop, nil, nil, nil), sideTop, sideBottom)

	u.noteEntry = widget.NewMultiLineEntry()
	u.noteEntry.TextStyle = fyne.TextStyle{Monospace: true}
	u.noteEntry.Wrapping = fyne.TextWrapWord

	// u.noteEntry.OnChanged = func(str string) { println(str) }
	return newAdaptiveSplit(side, u.noteEntry)
}

func (u *ui) showMarkdownPopup(parentCanvas fyne.Canvas) {
	widget.ShowPopUp(widget.NewRichTextFromMarkdown(u.current.text), parentCanvas)
}

func promptUserForBookDir(w fyne.Window) {
	var bookDirs []string

	// get a list of directories

	homePath := path.Join(theUserHomeDir, theDataDir)
	f, err := os.Open(homePath)
	if err != nil {
		log.Fatalf("couldn't open path %s: %s\n", homePath, err)
	}
	dirNames, err := f.Readdirnames(-1)
	if err != nil {
		log.Fatalf("couldn't read dir names for path %s: %s\n", homePath, err)
	}
	bookDirs = append(bookDirs, dirNames...)
	if len(bookDirs) == 1 {
		theBookDir = bookDirs[0]
		return
	}

	// magic up a list box

	// this looks fugly and opens up in the home directory and doesn't show hidden directories
	// dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
	// 	fmt.Println(uri)
	// }, w)

	selectedBook := theBookDir
	combo := widget.NewSelect(bookDirs, func(book string) {
		selectedBook = book
	})
	dialog.ShowCustomConfirm("Select Book", "OK", "Cancel", combo, func(ok bool) {
		if ok && theBookDir != selectedBook {
			theUI.saveDirtyNote()
			theUI.found = []*note{}
			theUI.foundList.Refresh()
			fmt.Println("setting theBookDir to", theBookDir)
			theBookDir = selectedBook
			theUI.setCurrent(load(time.Now()))
			w.SetTitle("Gold Notebook - " + theBookDir)
		}
	}, w)

	/*
	   var selectedDir int
	   lbox := widget.NewList(

	   	func() int {
	   		return len(bookDirs)
	   	},
	   	func() fyne.CanvasObject {
	   		return widget.NewLabel("")
	   	},
	   	func(id widget.ListItemID, obj fyne.CanvasObject) {
	   		// println("update widget.ListItemID", id)
	   		obj.(*widget.Label).SetText(bookDirs[id])
	   	},

	   )

	   	lbox.OnSelected = func(id int) {
	   		selectedDir = id
	   	}

	   lbox.UnselectAll()

	   	for i, d := range bookDirs {
	   		if d == theBookDir {
	   			lbox.Select(i)
	   			lbox.ScrollTo(i)
	   			break
	   		}
	   	}

	   fmt.Println(lbox.MinSize()) // approx 32, 33
	   // content := container.New(&DirListLayout{}, lbox)

	   	dialog.ShowCustomConfirm("Select Book", "OK", "Cancel", lbox, func(ok bool) {
	   		if !ok {
	   			selectedDir = -1
	   			fmt.Println("leaving theBookDir untouched")
	   			return
	   		} else {
	   			theBookDir = bookDirs[selectedDir]
	   			fmt.Println("setting theBookDir to", theBookDir)
	   			return
	   		}
	   	}, w)
	*/
}

func main() {
	{
		var err error
		theUserHomeDir, err = os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
	}
	flag.BoolVar(&debugMode, "debug", false, "turn debug mode on")
	flag.StringVar(&theDataDir, "data", ".goldnotebook", "name of the data directory")
	flag.StringVar(&theBookDir, "book", "Common", "name of the book to open")
	flag.Parse()
	// if len(flag.Args()) == 1 {
	// 	theBookDir = flag.Args()[0]
	// }

	if debugMode {
		fmt.Println("home:", theUserHomeDir, "\ndata:", theDataDir, "\nbook:", theBookDir)
	}

	a := app.NewWithID("oddstream.goldnotebook")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "book-48.png",
		StaticContent: book48IconBytes,
	})
	a.Settings().SetTheme(&noteTheme{})

	// if theBookDir is set on command line, don't ask the user for it

	w := a.NewWindow("Gold Notebook - " + theBookDir)

	theUI = &ui{current: load(time.Now())}

	// shortcuts get swallowed if focus is in the note multiline entry widget
	ctrlF := &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		w.Canvas().Focus(theUI.searchEntry)
	})
	ctrlM := &desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlM, func(shortcut fyne.Shortcut) {
		theUI.showMarkdownPopup(w.Canvas())
	})
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theUI.saveDirtyNote()
	})
	ctrlB := &desktop.CustomShortcut{KeyName: fyne.KeyB, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlB, func(shortcut fyne.Shortcut) {
		promptUserForBookDir(w)
	})
	w.SetContent(buildUI(theUI))
	w.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theUI.current.text)

	w.Resize(fyne.NewSize(1024, 640))
	w.CenterOnScreen()
	w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theUI.saveDirtyNote()
}
