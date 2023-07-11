package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	"oddstream.cj/fynex"
	"oddstream.cj/note"
	"oddstream.cj/util"
)

//go:embed today-128.png
var todayIconBytes []byte // https://www.iconsdb.com/white-icons/today-icon.html

const (
	appName    = "cj"
	appVersion = "0.1"
)

type cjNote struct {
	note.Note
	date time.Time
}

func (n *cjNote) daysBetween(m *cjNote) int {
	return int(n.date.Sub(m.date).Hours() / 24)
}

func (n *cjNote) fname() string {
	return path.Join(theUserHomeDir, theDataDir, theJournalDir,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()),
		fmt.Sprintf("%02d.txt", n.date.Day()))
}

func makeAndLoadNote(t time.Time) *cjNote {
	n := &cjNote{date: t}
	fname := n.fname()
	n.Load(fname)
	return n
}

var (
	theUI          *ui
	theUserHomeDir string // eg /home/gilbert
	theDataDir     string // eg .cj
	theJournalDir  string // eg Default
	debugMode      bool
)

type ui struct {
	current *cjNote
	found   []*cjNote

	w           fyne.Window
	toolbar     *widget.Toolbar
	calendar    *fyne.Container //*Calendar
	searchEntry *widget.Entry
	foundList   *widget.List
	noteEntry   *widget.Entry
}

func appTitle() string {
	return "Commonplace Journal - " + theJournalDir
}

func parseDateFromFname(pathname string) time.Time {
	// TODO this is as ugly as fuck and needs reworking
	// pathname comes from the output of grep, and looks like
	// /home/gilbert/.cj/Default/2023/06/10.txt
	var t = time.Time{}
	lst := strings.Split(pathname, string(os.PathSeparator))
	for i := 0; i < len(lst)-3; i++ {
		if lst[i] == theJournalDir {
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

func (u *ui) setCurrent(n *cjNote) {
	u.current = n
	u.noteEntry.SetText(theUI.current.Text)
	u.calendar.Objects[0] = fynex.NewCalendar(theUI.current.date, calendarTapped, calendarIsDateImportant)
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

func find(query string) []*cjNote {
	var found []*cjNote

	if query == "" {
		return found
	}

	cmd := exec.Command("grep",
		"--extended-regexp",
		"--recursive",
		"--ignore-case",
		"--files-with-matches",
		regexp.QuoteMeta(query),
		path.Join(theUserHomeDir, theDataDir))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with %s\n", err)
	}
	stdin := bufio.NewScanner(stdout)
	for stdin.Scan() {
		// fmt.Println(stdin.Text())
		n := makeAndLoadNote(parseDateFromFname(stdin.Text()))
		found = append(found, n)
	}
	cmd.Wait()

	return found
}

func listSelected(id widget.ListItemID) {
	// log.Printf("list item %d selected", id)
	theUI.saveDirtyNote()
	theUI.setCurrent(theUI.found[id])
}

func contains(lst []*cjNote, b *cjNote) bool {
	for _, n := range lst {
		if n.daysBetween(b) == 0 {
			return true
		}
	}
	return false
}

func (u *ui) postFind(query string) {
	if len(u.found) > 0 {
		u.searchEntry.Text = query
		u.searchEntry.Refresh()

		u.foundList.Select(0)
		u.foundList.Refresh()

		u.setCurrent(u.found[0])
	} else {
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
}

func (u *ui) findEx() {
	var pu *widget.PopUp

	ent := widget.NewEntry()
	ent.PlaceHolder = "Search"
	widen := widget.NewButton("Widen", func() {
		results := find(ent.Text)
		if len(results) == 0 {
			return
		}
		var newFound []*cjNote = u.found
		for _, n := range results {
			if !contains(u.found, n) {
				newFound = append(newFound, n)
			}
		}
		u.found = newFound
		u.postFind("")
		pu.Hide()
	})
	narrow := widget.NewButton("Narrow", func() {
		results := find(ent.Text)
		if len(results) == 0 {
			return
		}
		var newFound []*cjNote
		for _, n := range results {
			if contains(u.found, n) {
				newFound = append(newFound, n)
			}
		}
		u.found = newFound
		u.postFind("")
		pu.Hide()
	})
	exclude := widget.NewButton("Exclude", func() {
		results := find(ent.Text)
		if len(results) == 0 {
			return
		}
		var newFound []*cjNote
		for _, n := range results {
			if !contains(u.found, n) {
				newFound = append(newFound, n)
			}
		}
		u.found = newFound
		u.postFind("")
		pu.Hide()
	})
	cancel := widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	content := container.New(layout.NewVBoxLayout(), ent, widen, narrow, exclude, cancel)

	pu = widget.NewModalPopUp(content, u.w.Canvas())
	pu.Show()
}

func buildUI(u *ui) fyne.CanvasObject {
	u.toolbar = widget.NewToolbar(
		// https://developer.fyne.io/explore/icons
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			u.promptUserForJournalDir()
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
		u.found = []*cjNote{}
		if len(str) > 1 {
			u.found = find(str)
		}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
	// u.searchEntry.OnSubmitted = func(str string) {
	// 	u.found = find(str)
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

func (u *ui) promptUserForJournalDir() {
	var journalDirs []string

	homePath := path.Join(theUserHomeDir, theDataDir)
	f, err := os.Open(homePath)
	if err != nil {
		log.Fatalf("couldn't open path %s: %s\n", homePath, err)
	}
	dirNames, err := f.Readdirnames(-1)
	if err != nil {
		log.Fatalf("couldn't read dir names for path %s: %s\n", homePath, err)
	}
	for _, dirName := range dirNames {
		if !strings.HasPrefix(dirName, ".") {
			journalDirs = append(journalDirs, dirName)
		}
	}
	if len(journalDirs) == 1 {
		theJournalDir = journalDirs[0]
		return
	}

	// this looks fugly and opens up in the home directory and doesn't show hidden directories
	// dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
	// 	fmt.Println(uri)
	// }, w)

	fynex.ShowListEntryPopUp2(u.w.Canvas(), "Select Journal", journalDirs, func(str string) {
		if str == "" {
			return
		}
		if str != theJournalDir {
			theJournalDir = str
			calendarTapped(time.Now())
			u.found = []*cjNote{}
			u.foundList.Refresh()
			u.w.SetTitle(appTitle())
		}
	})
}

func (u *ui) searchForHashTags() {
	cmd := exec.Command("grep",
		"--extended-regexp",
		"--recursive",
		"--ignore-case",
		"--only-matching",
		"--no-filename",
		"#[[:alnum:]]+",
		path.Join(theUserHomeDir, theDataDir))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with %s\n", err)
	}
	stdin := bufio.NewScanner(stdout)
	var results []string
	for stdin.Scan() {
		txt := stdin.Text()
		txt = strings.ToLower(txt)
		// fmt.Println(txt)
		results = append(results, txt)
	}
	cmd.Wait()
	if len(results) > 0 {
		results = util.RemoveDuplicateStrings(results)
		fynex.ShowListPopUp2(theUI.w.Canvas(), "Find Hashtag", results, func(str string) {
			theUI.found = find(str)
			theUI.postFind(str)
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
	var windowWidth, windowHeight int
	reportVersion := flag.Bool("version", false, "report app version")
	flag.BoolVar(&debugMode, "debug", false, "turn debug mode on")
	flag.StringVar(&theDataDir, "data", ".cj", "name of the data directory")
	flag.StringVar(&theJournalDir, "Journal", "Default", "name of the journal to open")
	flag.IntVar(&windowWidth, "width", 1024, "width of the window")
	flag.IntVar(&windowHeight, "height", 640, "height of the window")
	flag.Parse()
	if *reportVersion {
		fmt.Println(appName, appVersion)
		os.Exit(0)
	}

	if debugMode {
		// if str, err := os.Executable(); err != nil {
		// 	log.Printf("err: %T, %v\n", err, err)
		// } else {
		// 	log.Printf("str: %T, %v\n", str, str)
		// }
		log.Println("\nhome:", theUserHomeDir, "\ndata:", theDataDir, "\njournal:", theJournalDir)
	}

	a := app.NewWithID("oddstream.cj")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "today.png",
		StaticContent: todayIconBytes,
	})

	a.Settings().SetTheme(fynex.NewNoteTheme(path.Join(theUserHomeDir, theDataDir, "theme.json")))

	theUI = &ui{w: a.NewWindow(appTitle()), current: makeAndLoadNote(time.Now())}

	// shortcuts get swallowed if focus is in the note multiline entry widget
	ctrlF := &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		if len(theUI.found) > 0 {
			theUI.findEx()
		}
	})
	// don't need this: just tap the 'today' icon in the taskbar
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theUI.saveDirtyNote()
	})

	theUI.w.SetContent(buildUI(theUI))
	theUI.w.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theUI.current.Text)

	theUI.w.Resize(fyne.NewSize(float32(windowWidth), float32(windowHeight)))
	theUI.w.CenterOnScreen()
	theUI.w.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theUI.saveDirtyNote()
}
