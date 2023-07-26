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
	"path/filepath"
	"regexp"
	"sort"
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

//go:embed today-48.png
var todayIconBytes []byte // https://www.iconsdb.com/white-icons/today-icon.html

const (
	appName    = "cj"
	appVersion = "0.1"
)

type cjNote struct {
	note.Note
	date     time.Time
	pathname string
}

func makeAndLoadNote(date time.Time) *cjNote {
	n := &cjNote{date: date}
	n.pathname = path.Join(theUserHomeDir, theDataDir, theJournalDir,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()),
		fmt.Sprintf("%02d.txt", n.date.Day()))
	n.Load(n.pathname)
	return n
}

func (n *cjNote) daysBetween(m *cjNote) int {
	return int(n.date.Sub(m.date).Hours() / 24)
}

var (
	theUI          *ui
	theUserHomeDir string    // eg /home/gilbert
	theDataDir     string    // eg .cj
	theJournalDir  string    // eg Default
	theNote        *cjNote   // the current note
	theFound       []*cjNote // the list of found notes
	debugMode      bool
)

type ui struct {
	mainWindow  fyne.Window
	toolbar     *widget.Toolbar
	calendar    *fyne.Container //*Calendar
	searchEntry *widget.Entry
	foundList   *widget.List
	noteEntry   *widget.Entry
	theme       fyne.Theme // Theme is an interface
}

func appTitle(t time.Time) string {
	return "Commonplace Journal - " + theJournalDir
	// return t.Format("Mon 2 Jan 2006") + " - " + theJournalDir
}

func saveDirtyNote() {
	newText := theUI.noteEntry.Text
	if newText != theNote.Text {
		if util.IsStringEmpty(newText) {
			theNote.Remove(theNote.pathname)
		} else {
			theNote.Text = newText
			theNote.Save(theNote.pathname)
		}
	}
}

func (u *ui) setCurrent(n *cjNote) {
	theNote = n
	u.noteEntry.SetText(theNote.Text)
	u.calendar.Objects[0] = fynex.NewCalendar(theNote.date, calendarTapped, calendarIsDateImportant)
	u.mainWindow.SetTitle(appTitle(n.date))
}

func calendarTapped(t time.Time) {
	saveDirtyNote()
	theUI.setCurrent(makeAndLoadNote(t))
	theUI.foundList.UnselectAll()
	theUI.mainWindow.Canvas().Focus(theUI.noteEntry)
}

func calendarIsDateImportant(t time.Time) bool {
	return t.Year() == theNote.date.Year() &&
		t.Month() == theNote.date.Month() &&
		t.Day() == theNote.date.Day()
}

func parseDateFromFname(pathname string) time.Time {
	// TODO this is as ugly as fuck and needs reworking
	// pathname comes from the output of grep, and looks like
	// /home/gilbert/.cj/Default/2023/06/10.txt
	// so we know in advance what the prefix of the pathname is,
	// so remove it to get
	// 2023/06/10.txt
	// then remove the suffix to get
	// 2023/06/10
	{
		var t time.Time = time.Time{}

		prefix := path.Join(theUserHomeDir, theDataDir, theJournalDir) + "/"
		str := strings.TrimPrefix(pathname, prefix)
		ext := filepath.Ext(str) // includes .
		str = strings.TrimSuffix(str, ext)
		lst := strings.Split(str, string(os.PathSeparator))
		if len(lst) == 3 {
			y, _ := strconv.Atoi(lst[0])
			m, _ := strconv.Atoi(lst[1])
			d, _ := strconv.Atoi(lst[2])
			t = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
		}
		// fmt.Println(pathname, y, m, d)
		/*
			psattern := fmt.Sprintf("/%c2006%c01%c02",
				os.PathSeparator,
				os.PathSeparator,
				os.PathSeparator)
			// fmt.Println(t)
			t2, err := time.Parse(psattern, str)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(t2)
			}
		*/
		return t
	}
	/*
	   var t = time.Time{}
	   lst := strings.Split(pathname, string(os.PathSeparator))

	   	for i := 0; i < len(lst)-3; i++ {
	   		if lst[i] == theJournalDir {
	   			if y, err := strconv.Atoi(lst[i+1]); err == nil {
	   				if m, err := strconv.Atoi(lst[i+2]); err == nil {
	   					if f, _, ok := strings.Cut(lst[i+3], "."); ok {
	   						if d, err := strconv.Atoi(f); err == nil {
	   							t = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
	   							break
	   						}
	   					}
	   				}
	   			}
	   		}
	   	}

	   return t
	*/
}

func find(query string) []*cjNote {
	var found []*cjNote

	if query == "" {
		return found
	}

	// could use xargs to run several directories in parallel?
	// expected output is a list of pathnames, one per line, eg
	// /home/gilbert/.cj/Default/2023/07/04.txt
	// /home/gilbert/.cj/Default/2023/06/18.txt
	// could use ripgrep which is faster
	cmd := exec.Command("grep",
		"--fixed-strings", // Interpret PATTERNS as fixed strings, not regular expressions.
		"--recursive",
		"--ignore-case",
		"--files-with-matches", // print the name of each input file from which output would normally have been printed.
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
		n := &cjNote{pathname: stdin.Text()}
		n.date = parseDateFromFname(n.pathname)
		n.Load(n.pathname)
		found = append(found, n)
	}
	cmd.Wait() // ignore error return because we're done

	sort.Slice(found, func(i, j int) bool {
		return found[i].date.Before(found[j].date)
	})

	return found
}

func contains(lst []*cjNote, b *cjNote) bool {
	for _, n := range lst {
		if n.daysBetween(b) == 0 {
			return true
		}
	}
	return false
}

func (u *ui) postFind() {
	if len(theFound) > 0 {
		u.foundList.Select(0)
		u.setCurrent(theFound[0])
	} else {
		u.foundList.UnselectAll()
	}
	u.foundList.Refresh()
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
		var newFound []*cjNote = theFound
		for _, n := range results {
			if !contains(theFound, n) {
				newFound = append(newFound, n)
			}
		}
		theFound = newFound
		u.postFind()
		pu.Hide()
	})
	narrow := widget.NewButton("Narrow", func() {
		results := find(ent.Text)
		if len(results) == 0 {
			return
		}
		var newFound []*cjNote
		for _, n := range results {
			if contains(theFound, n) {
				newFound = append(newFound, n)
			}
		}
		theFound = newFound
		u.postFind()
		pu.Hide()
	})
	exclude := widget.NewButton("Exclude", func() {
		results := find(ent.Text)
		if len(results) == 0 {
			return
		}
		var newFound []*cjNote
		for _, n := range results {
			if !contains(theFound, n) {
				newFound = append(newFound, n)
			}
		}
		theFound = newFound
		u.postFind()
		pu.Hide()
	})
	cancel := widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	content := container.New(layout.NewVBoxLayout(), ent, widen, narrow, exclude, cancel)

	pu = widget.NewModalPopUp(content, u.mainWindow.Canvas())
	pu.Show()
	pu.Canvas.Focus(ent)
}

func buildUI(u *ui) fyne.CanvasObject {
	u.toolbar = widget.NewToolbar(
		// https://developer.fyne.io/explore/icons
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			u.promptUserForJournalDir()
		}),
		// widget.NewToolbarAction(theme.SearchIcon(), func() {
		widget.NewToolbarAction(u.theme.Icon("tag"), func() {
			theUI.searchForHashTags()
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			// if len(theFound) > 0 {
			theUI.findEx()
			// }
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			t := theNote.date
			t = t.Add(-time.Hour * 24)
			calendarTapped(t)
		}),
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			calendarTapped(time.Now())
		}),
		widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
			t := theNote.date
			t = t.Add(time.Hour * 24)
			calendarTapped(t)
		}),
	)

	u.calendar = container.New(layout.NewCenterLayout(), fynex.NewCalendar(theNote.date, calendarTapped, calendarIsDateImportant))

	u.searchEntry = widget.NewEntry()
	u.searchEntry.PlaceHolder = "Search"
	u.searchEntry.OnChanged = func(str string) {
		theFound = []*cjNote{}
		if len(str) > 1 {
			theFound = find(str)
		}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	}
	// u.searchEntry.OnSubmitted = func(str string) {
	// 	u.found = find(str)
	// }
	u.searchEntry.TextStyle = fyne.TextStyle{Monospace: true}

	searchEntryClear := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		u.searchEntry.SetText("")
		theUI.mainWindow.Canvas().Focus(theUI.searchEntry)
		theFound = []*cjNote{}
		u.foundList.UnselectAll()
		u.foundList.Refresh()
	})

	u.foundList = widget.NewList(
		func() int {
			return len(theFound)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// obj.(*widget.Label).SetText(util.FirstLine(theFound[id].Text))
			obj.(*widget.Label).SetText(theFound[id].date.Format("Mon 2 Jan 2006"))
		},
	)
	u.foundList.OnSelected = func(id widget.ListItemID) {
		saveDirtyNote()
		theUI.setCurrent(theFound[id])
	}

	u.noteEntry = widget.NewMultiLineEntry()
	u.noteEntry.TextStyle = fyne.TextStyle{Monospace: true}
	u.noteEntry.Wrapping = fyne.TextWrapWord

	// https://developer.fyne.io/explore/layouts

	searchForm := container.New(layout.NewBorderLayout(nil, nil, nil, searchEntryClear), searchEntryClear, u.searchEntry)
	sideTop := container.New(layout.NewVBoxLayout(), u.calendar, searchForm)
	sideBottom := container.New(layout.NewMaxLayout(), u.foundList)
	side := container.New(layout.NewBorderLayout(sideTop, nil, nil, nil), sideTop, sideBottom)

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

	fynex.ShowListEntryPopUp2(u.mainWindow.Canvas(), "Select Journal", journalDirs, func(str string) {
		if str == "" {
			return
		}
		if str != theJournalDir {
			theJournalDir = str
			calendarTapped(time.Now())
			theFound = []*cjNote{}
			u.foundList.Refresh()
		}
	})
}

func (u *ui) searchForHashTags() {
	cmd := exec.Command("grep",
		"--extended-regexp", // because we are using character class
		"--recursive",
		"--ignore-case",
		"--only-matching",
		"--no-filename",
		"-I", // don't process binary files
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
	cmd.Wait() // ignore error return because we're done
	if len(results) > 0 {
		results = util.RemoveDuplicateStrings(results) // sorts slice as a side-effect
		fynex.ShowListPopUp2(theUI.mainWindow.Canvas(), "Find Hashtag", results, func(str string) {
			theFound = find(str)
			theUI.postFind()
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

	// theTheme := fynex.NewNoteTheme()
	// a.Settings().SetTheme(&theTheme)

	theUI = &ui{mainWindow: a.NewWindow(appTitle(time.Now())), theme: fynex.NewNoteTheme()}
	a.Settings().SetTheme(theUI.theme)
	theNote = makeAndLoadNote(time.Now())

	// shortcuts get swallowed if focus is in the note multiline entry widget
	// don't need this: just tap the 'today' icon in the taskbar
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.mainWindow.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		saveDirtyNote()
	})

	theUI.mainWindow.SetContent(buildUI(theUI))
	theUI.mainWindow.Canvas().Focus(theUI.noteEntry)
	theUI.noteEntry.SetText(theNote.Text)

	theUI.mainWindow.Resize(fyne.NewSize(float32(windowWidth), float32(windowHeight)))
	theUI.mainWindow.CenterOnScreen()
	theUI.mainWindow.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	saveDirtyNote()
}
