package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	appName    = "Nincomp"
	appVersion = "0.1"
)

type ui struct {
	current *incNote
	found   []*incNote

	w           fyne.Window
	toolbar     *widget.Toolbar
	calendar    *fyne.Container //*Calendar
	searchEntry *widget.Entry
	foundList   *widget.List
	noteEntry   *widget.Entry
}

type incNote struct {
	note.Note
	date time.Time
}

func (n *incNote) fname() string {
	return path.Join(theUserHomeDir, theDataDir, "inc", theBookDir,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()),
		fmt.Sprintf("%02d.txt", n.date.Day()))
}

func makeAndLoadNote(t time.Time) *incNote {
	n := &incNote{date: t}
	fname := n.fname()
	n.Load(fname)
	return n
}

var (
	theUI          *ui
	theUserHomeDir string // eg /home/gilbert
	theDataDir     string // eg .nincomp
	theBookDir     string // eg Default
	debugMode      bool
)

func appTitle() string {
	return "Incremental Notes - " + theBookDir
}

func parseDateFromFname(fname string) time.Time {
	var t = time.Time{}
	lst := strings.Split(fname, string(os.PathSeparator))
	for i := 0; i < len(lst)-3; i++ {
		if lst[i] == theBookDir {
			if y, err := strconv.Atoi(lst[i+1]); err == nil {
				if m, err := strconv.Atoi(lst[i+2]); err == nil {
					if f, _, ok := strings.Cut(lst[i+3], "."); ok {
						if d, err := strconv.Atoi(f); err == nil {
							t = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
							break
						}
					}
				}
			}
		}
	}
	return t
}

func (u *ui) saveDirtyNote() {
	newText := u.noteEntry.Text
	if newText != u.current.Text {
		u.current.Text = newText
		u.current.Save(u.current.fname())
	}
}

func (u *ui) setCurrent(n *incNote) {
	theUI.current = n
	theUI.noteEntry.SetText(theUI.current.Text)
	theUI.calendar.Objects[0] = fynex.NewCalendar(theUI.current.date, calendarTapped, calendarIsDateImportant)
}

func calendarTapped(t time.Time) {
	theUI.saveDirtyNote()
	theUI.setCurrent(makeAndLoadNote(t))
	theUI.foundList.UnselectAll()
}

func calendarIsDateImportant(t time.Time) bool {
	return t.Year() == theUI.current.date.Year() &&
		t.Month() == theUI.current.date.Month() &&
		t.Day() == theUI.current.date.Day()
}

func (u *ui) find(query string) {
	if query == "" {
		return
	}
	// query = strings.ToLower(query)

	u.found = []*incNote{}

	opts := &search.SearchOptions{
		Kind:   search.LITERAL,
		Regex:  nil,
		Finder: search.MakeStringFinder([]byte(query)),
	}
	results := search.Search([]string{path.Join(theUserHomeDir, theDataDir, "inc", theBookDir)}, opts)
	for _, fname := range results {
		n := makeAndLoadNote(parseDateFromFname(fname))
		u.found = append(theUI.found, n)
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
			u.promptUserForBookDir()
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			theUI.searchForHashTags()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			t := theUI.current.date
			t = t.Add(-time.Hour * 24)
			calendarTapped(t)
		}),
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			calendarTapped(time.Now())
		}),
		widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
			t := theUI.current.date
			t = t.Add(time.Hour * 24)
			calendarTapped(t)
		}),
	)
	u.calendar = container.New(layout.NewCenterLayout(), fynex.NewCalendar(theUI.current.date, calendarTapped, calendarIsDateImportant))
	u.searchEntry = widget.NewEntry()
	u.searchEntry.PlaceHolder = "Search"
	u.searchEntry.OnChanged = func(str string) {
		u.found = []*incNote{}
		if len(str) > 1 {
			u.find(str)
		}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
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
			obj.(*widget.Label).SetText(util.FirstLine(theUI.found[id].Text))
		},
	)
	u.foundList.OnSelected = listSelected

	sideTop := container.New(layout.NewVBoxLayout(), u.calendar, u.searchEntry)
	sideBottom := container.New(layout.NewMaxLayout(), u.foundList)
	side := container.New(layout.NewBorderLayout(sideTop, nil, nil, nil), sideTop, sideBottom)

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

	homePath := path.Join(theUserHomeDir, theDataDir, "inc")
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
		u.found = []*incNote{}
		u.foundList.Refresh()
		// if debugMode {
		// 	log.Println("setting theBookDir to", theBookDir)
		// }
		theBookDir = str
		u.setCurrent(&incNote{})
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
	results := search.Search([]string{path.Join(theUserHomeDir, theDataDir, "inc", theBookDir)}, opts)
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
	var startSearch string // so com and inc have same command line flags
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

	a := app.NewWithID("oddstream.incrementalnotebook")

	th := &fynex.NoteTheme{}
	a.Settings().SetTheme(th)
	a.SetIcon(th.BookIcon())

	theUI = &ui{w: a.NewWindow(appTitle()), current: makeAndLoadNote(time.Now())}

	// shortcuts get swallowed if focus is in the note multiline entry widget
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theUI.saveDirtyNote()
	})

	theUI.w.SetContent(buildUI(theUI))
	theUI.w.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theUI.current.Text)

	// if startSearch != "" {
	// 	if !strings.HasPrefix(startSearch, "#") {
	// 		startSearch = "#" + startSearch
	// 	}
	// 	theUI.injectSearch(startSearch)
	// }

	theUI.w.Resize(fyne.NewSize(1024, 640))
	theUI.w.CenterOnScreen()
	theUI.w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theUI.saveDirtyNote()
}
