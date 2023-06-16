package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"oddstream.nincomp/fynex"
	"oddstream.nincomp/note"
	"oddstream.nincomp/search"
	"oddstream.nincomp/util"
)

const (
	appName    = "com"
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
	ext   string // the extension of this note WHEN LOADED (usually .txt or .md)
}

func load(fname string) *comNote {
	cn := &comNote{}
	cn.Load(fname)
	cn.title = util.FirstLine(cn.Text)
	cn.ext = path.Ext(fname) // DOES include leading .
	return cn
}

func (cn *comNote) save() {
	title := util.Sanitize(cn.title)
	if title == "" {
		title = "untitled"
	}
	fname := path.Join(theUserHomeDir, theDataDir, "com", theBookDir, title+cn.ext)
	cn.Save(fname)
}

func (cn *comNote) remove() {
	title := util.Sanitize(cn.title)
	if title == "" {
		title = "untitled"
	}
	fname := path.Join(theUserHomeDir, theDataDir, "com", theBookDir, title+cn.ext)
	cn.Remove(fname)
}

var (
	theUI          *ui
	theUserHomeDir string // eg /home/gilbert
	theDataDir     string // eg .noncomp
	theBookDir     string // eg Default
	debugMode      bool
)

func appTitle() string {
	return "Commonplace Book - " + theBookDir
}

func (u *ui) saveDirtyNote() {
	newText := u.noteEntry.Text
	if newText != u.current.Text {
		oldTitle := u.current.title
		newTitle := util.FirstLine(newText)
		if newTitle != oldTitle {
			u.current.remove()
		}
		u.current.Text = newText
		u.current.title = newTitle
		// keep the same .ext
		u.current.save()
		u.foundList.Refresh()
	}
}

func (u *ui) setCurrent(n *comNote) {
	theUI.current = n
	theUI.noteEntry.SetText(theUI.current.Text)
	theUI.w.Canvas().Focus(theUI.noteEntry)
}

// find takes the query, does a search, and fills the .found slice of comNotes
func (u *ui) find(query string) {
	if query == "" {
		return
	}

	u.found = []*comNote{}

	opts := &search.SearchOptions{
		Kind:   search.LITERAL,
		Regex:  nil,
		Finder: search.MakeStringFinder([]byte(query)),
	}
	results := search.Search([]string{path.Join(theUserHomeDir, theDataDir, "com", theBookDir)}, opts)
	for _, fname := range results {
		n := load(fname)
		u.found = append(u.found, n)
	}
}

// findAll fills the .found slice with all the notes in the current book
func (u *ui) findAll() {
	u.found = []*comNote{}

	bookDir := path.Join(theUserHomeDir, theDataDir, "com", theBookDir)
	files, err := os.ReadDir(bookDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// log.Println(path.Join(bookDir, file.Name()))
		n := load(path.Join(bookDir, file.Name()))
		u.found = append(u.found, n)
	}
	u.searchEntry.Text = ""
	u.searchEntry.Refresh()

	u.foundList.UnselectAll()
	u.foundList.Refresh()

	// for i, n := range u.found {
	// 	println(i, n.title)
	// }
}

func buildUI(u *ui) fyne.CanvasObject {
	u.toolbar = widget.NewToolbar(
		// https://developer.fyne.io/explore/icons
		// widget.NewToolbarAction(theme.MenuIcon(), func() {
		// 	fynex.ShowMenuPopup(theUI.w.Canvas(), u.toolbar)
		// }),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			theUI.promptUserForBookDir()
		}),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			theUI.saveDirtyNote()
			theUI.found = []*comNote{}
			theUI.foundList.Refresh()
			theUI.setCurrent(&comNote{})
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			theUI.searchForHashTags()
		}),
	)

	u.searchEntry = widget.NewEntry()
	u.searchEntry.OnChanged = func(str string) {
		u.found = []*comNote{}
		if len(str) > 1 {
			u.find(str)
		}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
	u.searchEntry.PlaceHolder = "Search"
	u.searchEntry.TextStyle = fyne.TextStyle{Monospace: true}

	u.foundList = widget.NewList(
		func() int {
			return len(theUI.found)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// println("update widget.ListItemID", id, theUI.found[id].title)
			obj.(*widget.Label).SetText(theUI.found[id].title)
		},
	)
	u.foundList.OnSelected = func(id widget.ListItemID) {
		theUI.saveDirtyNote()
		theUI.setCurrent(theUI.found[id])
	}

	side := container.New(layout.NewBorderLayout(u.searchEntry, nil, nil, nil), u.searchEntry, u.foundList)

	u.noteEntry = widget.NewMultiLineEntry()
	u.noteEntry.TextStyle = fyne.TextStyle{Monospace: true}
	u.noteEntry.Wrapping = fyne.TextWrapWord

	mainPanel := container.New(layout.NewBorderLayout(u.toolbar, nil, nil, nil), u.toolbar, u.noteEntry)

	// u.noteEntry.OnChanged = func(str string) { println(str) }
	return fynex.NewAdaptiveSplit(side, mainPanel)
}

func (u *ui) promptUserForBookDir() {
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

	// this looks fugly and opens up in the home directory and doesn't show hidden directories
	// dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
	// 	fmt.Println(uri)
	// }, w)

	fynex.ShowListEntryPopUp(u.w.Canvas(), "Select Book", bookDirs, func(str string) {
		if str == "" {
			return
		}
		u.saveDirtyNote()
		u.searchEntry.Text = ""
		u.searchEntry.Refresh()
		u.found = []*comNote{}
		u.foundList.Refresh()
		// if debugMode {
		// 	log.Println("setting theBookDir to", theBookDir)
		// }
		theBookDir = str
		u.setCurrent(&comNote{})
		u.w.SetTitle(appTitle())
	})
}

func (u *ui) injectSearch(query string) {
	u.find(query)
	if len(theUI.found) > 0 {
		u.searchEntry.Text = query
		u.searchEntry.Refresh()

		u.foundList.Select(0)
		u.foundList.Refresh()

		u.setCurrent(u.found[0])
	}
}

func (u *ui) searchForHashTags() {
	opts := &search.SearchOptions{
		Kind:   search.REGEX,
		Regex:  regexp.MustCompile("#[[:alnum:]]+"),
		Finder: nil,
	}
	results := search.Search([]string{path.Join(theUserHomeDir, theDataDir, "com", theBookDir)}, opts)
	if len(results) > 0 {
		fynex.ShowListPopUp2(theUI.w.Canvas(), "Find Hashtag", results, func(str string) {
			u.injectSearch(str)
		})
	}
}

func main() {
	{
		var err error
		theUserHomeDir, err = os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
	}
	var startSearch string
	reportVersion := flag.Bool("version", false, "report app version")
	flag.BoolVar(&debugMode, "debug", false, "turn debug mode on")
	flag.StringVar(&theDataDir, "data", ".nincomp", "name of the data directory")
	flag.StringVar(&theBookDir, "book", "Default", "name of the book to open")
	flag.StringVar(&startSearch, "search", "", "look for this hashtag when starting")
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

	th := &fynex.NoteTheme{FontSize: 15.0, IconName: "book"}
	a.Settings().SetTheme(th)
	a.SetIcon(th.BookIcon())

	theUI = &ui{w: a.NewWindow(appTitle()), current: &comNote{}} // start with an empty note

	// shortcuts get swallowed if focus is in the note multiline entry widget
	ctrlF5 := &desktop.CustomShortcut{KeyName: fyne.KeyF5, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlF5, func(shortcut fyne.Shortcut) {
		theUI.findAll()
	})
	ctrlM := &desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlM, func(shortcut fyne.Shortcut) {
		fynex.ShowMarkdownPopup(theUI.w.Canvas(), theUI.current.Text)
	})

	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theUI.saveDirtyNote()
	})
	// "global shortcuts don’t work when a shortcutable widget is focused"
	// to add shortcuts to Shortcutable widgets, see
	// https://developer.fyne.io/explore/shortcuts
	// https://github.com/fyne-io/fyne/issues/2627
	// func (e *Entry) TypedShortcut(shortcut fyne.Shortcut)
	theUI.w.SetContent(buildUI(theUI))
	theUI.w.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theUI.current.Text)

	if startSearch != "" {
		if !strings.HasPrefix(startSearch, "#") {
			startSearch = "#" + startSearch
		}
		theUI.injectSearch(startSearch)
	}

	theUI.w.Resize(fyne.NewSize(1024, 640))
	theUI.w.CenterOnScreen()
	theUI.w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theUI.saveDirtyNote()
}
