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
	"oddstream.goldnotebook/fynex"
	"oddstream.goldnotebook/note"
	"oddstream.goldnotebook/search"
	"oddstream.goldnotebook/util"
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

func makeFnameFromTitle(title string) string {
	title = util.Sanitize(title)
	if title == "" {
		title = "untitled"
	}
	fname := path.Join(theUserHomeDir, theDataDir, "com", theBookDir, title+".txt")
	// println("makeFnameFromTitle title:", title, "fname:", fname)
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
		newTitle := util.FirstLine(newText)
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
	theUI.w.Canvas().Focus(theUI.noteEntry)
}

func (u *ui) find(query string) {
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
		n.title = util.FirstLine(n.Text)
		u.found = append(u.found, n)
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
			theUI.promptUserForBookDir()
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			theUI.saveDirtyNote()
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
	// u.searchEntry.OnSubmitted = func(str string) {
	// 	u.find(str)
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

	// magic up a list box

	// this looks fugly and opens up in the home directory and doesn't show hidden directories
	// dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
	// 	fmt.Println(uri)
	// }, w)

	fynex.ShowListEntryPopUp(u.w.Canvas(), "Select Book", bookDirs, func(str string) {
		if str == "" {
			return
		}
		u.saveDirtyNote()
		u.found = []*comNote{}
		u.foundList.Refresh()
		// if debugMode {
		// 	log.Println("setting theBookDir to", theBookDir)
		// }
		theBookDir = str
		u.setCurrent(&comNote{})
		u.w.SetTitle(appTitle())
	})

	// widget.List is displayed in it's MinSize, which only displays one line...
	// tried wrapping layout and list, which didn't fix it
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
		fynex.ShowListPopUp(theUI.w.Canvas(), "Find Hashtag", results, func(str string) {
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
	flag.StringVar(&theDataDir, "data", ".goldnotebook", "name of the data directory")
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

	th := &fynex.NoteTheme{}
	a.Settings().SetTheme(th)
	a.SetIcon(th.BookIcon())

	theUI = &ui{w: a.NewWindow(appTitle()), current: &comNote{}} // start with an empty note

	// shortcuts get swallowed if focus is in the note multiline entry widget
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
