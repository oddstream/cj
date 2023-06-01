package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"oddstream.goldnotebook/fynex"
	"oddstream.goldnotebook/note"
	"oddstream.goldnotebook/search"
)

const (
	appName    = "Goldnotebook"
	appVersion = "0.1"
)

type ui struct {
	current *comNote
	found   []*comNote

	w           fyne.Window
	toolbar     *widget.Toolbar
	searchEntry *widget.Entry
	foundList   *widget.List
	noteEntry   *widget.Entry

	// popUp        *widget.PopUp
}

type comNote struct {
	note.Note
	title string // the title of this note WHEN LOADED (to detect filename changes)
}

func sanitize(str string) string {
	var b strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune(' ')
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	s = strings.TrimSpace(s)
	return s
}

func firstLine(text string) string {
	var line string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line = scanner.Text()
		if len(line) > 0 {
			break
		}
	}
	return line
}

func makeFnameFromTitle(title string) string {
	title = sanitize(title)
	if title == "" {
		title = "untitled"
	}
	fname := path.Join(theUserHomeDir, theDataDir, "com", theBookDir, title+".txt")
	println("makeFnameFromTitle title:", title, "fname:", fname)
	return fname
}

var (
	theUI          *ui
	theUserHomeDir string // eg /home/gilbert
	theDataDir     string // eg .goldnotebook
	theBookDir     string // eg Common
	debugMode      bool
)

func appTitle() string {
	return "Commonplace Book - " + theBookDir
}

func (u *ui) saveDirtyNote() {
	newText := u.noteEntry.Text
	if newText != u.current.Text {
		oldTitle := u.current.title
		newTitle := firstLine(newText)
		u.current.Text = newText
		u.current.title = newTitle
		u.current.Save(makeFnameFromTitle(newTitle))
		if newTitle != oldTitle {
			u.current.Remove(makeFnameFromTitle(oldTitle))
		}
		u.foundList.Refresh()
	}
}

func (u *ui) setCurrent(n *comNote) {
	theUI.current = n
	theUI.noteEntry.SetText(theUI.current.Text)
}

func find(query string) {
	if query == "" {
		return
	}
	// query = strings.ToLower(query)

	theUI.found = []*comNote{}

	opts := &search.SearchOptions{
		Kind:   search.LITERAL,
		Regex:  nil,
		Finder: search.MakeStringFinder([]byte(query)),
	}
	results := search.Search([]string{path.Join(theUserHomeDir, theDataDir, "com", theBookDir)}, opts)
	for _, fname := range results {
		n := &comNote{}
		n.Load(fname)
		n.title = firstLine(n.Text)
		theUI.found = append(theUI.found, n)
	}
}

func listSelected(id widget.ListItemID) {
	// log.Printf("list item %d selected", id)
	theUI.saveDirtyNote()
	theUI.setCurrent(theUI.found[id])
}

func buildUI(u *ui) fyne.CanvasObject {
	u.toolbar = widget.NewToolbar(
		// https://developer.fyne.io/explore/icons
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			promptUserForBookDir()
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			theUI.saveDirtyNote()
			theUI.setCurrent(&comNote{})
		}),
	)
	u.searchEntry = widget.NewEntry()
	u.searchEntry.OnChanged = func(str string) {
		u.found = []*comNote{}
		if len(str) > 1 {
			find(str)
		}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
	// u.searchEntry.OnSubmitted = func(str string) {
	// 	find(str)
	// }
	u.searchEntry.TextStyle = fyne.TextStyle{Monospace: true}
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

	side := container.New(layout.NewBorderLayout(u.searchEntry, nil, nil, nil), u.searchEntry, u.foundList)

	u.noteEntry = widget.NewMultiLineEntry()
	u.noteEntry.TextStyle = fyne.TextStyle{Monospace: true}
	u.noteEntry.Wrapping = fyne.TextWrapWord

	mainPanel := container.New(layout.NewBorderLayout(u.toolbar, nil, nil, nil), u.toolbar, u.noteEntry)

	// u.noteEntry.OnChanged = func(str string) { println(str) }
	return fynex.NewAdaptiveSplit(side, mainPanel)
}

// func (u *ui) showMarkdownPopup(parentCanvas fyne.Canvas) {
// 	widget.ShowPopUp(widget.NewRichTextFromMarkdown(u.current.text), parentCanvas)
// }

func promptUserForBookDir() {
	var bookDirs []string

	// get a list of directories

	homePath := path.Join(theUserHomeDir, theDataDir, "com")
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

	rgroup := widget.NewRadioGroup(bookDirs, func(book string) {
		selectedBook = book
	})
	rgroup.Selected = theBookDir

	// sel := widget.NewSelect(bookDirs, func(str string) { selectedBook = str })

	entry := widget.NewEntry()
	content := container.New(layout.NewVBoxLayout(), rgroup, entry)
	dialog.ShowCustomConfirm("Select Book", "OK", "Cancel", content, func(ok bool) {
		if ok {
			if len(entry.Text) > 0 {
				selectedBook = entry.Text
			}
			if theBookDir != selectedBook {
				theUI.saveDirtyNote()
				theUI.found = []*comNote{}
				theUI.foundList.Refresh()
				if debugMode {
					log.Println("setting theBookDir to", theBookDir)
				}
				theBookDir = selectedBook
				theUI.setCurrent(&comNote{})
				theUI.w.SetTitle(appTitle())
			}
		}
	}, theUI.w)

	// widget.List is displayed in it's MinSize, which only displays one line...
	// tried wrapping layout and list, which didn't fix it

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
	reportVersion := flag.Bool("version", false, "report app version")
	flag.BoolVar(&debugMode, "debug", false, "turn debug mode on")
	flag.StringVar(&theDataDir, "data", ".goldnotebook", "name of the data directory")
	flag.StringVar(&theBookDir, "book", "Default", "name of the book to open")
	flag.Parse()
	if *reportVersion {
		fmt.Println(appName, appVersion)
		os.Exit(0)
	}

	if debugMode {
		if str, err := os.Executable(); err != nil {
			log.Printf("err: %T, %v\n", err, err)
		} else {
			log.Printf("str: %T, %v\n", str, str)
		}
		log.Println("\nhome:", theUserHomeDir, "\ndata:", theDataDir, "\nbook:", theBookDir)
	}

	a := app.NewWithID("oddstream.commonplacebook")

	th := &fynex.NoteTheme{}
	a.Settings().SetTheme(th)
	a.SetIcon(th.BookIcon())

	w := a.NewWindow(appTitle())
	theUI = &ui{w: w, current: &comNote{}} // start with an empty note

	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theUI.saveDirtyNote()
	})

	theUI.w.SetContent(buildUI(theUI))
	theUI.w.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theUI.current.Text)

	theUI.w.Resize(fyne.NewSize(1024, 640))
	theUI.w.CenterOnScreen()
	theUI.w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theUI.saveDirtyNote()
}
