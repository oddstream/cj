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
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

var (
	theUI          *ui
	theUserHomeDir string       // eg /home/gilbert
	theDataDir     string       // eg .cj
	theJournalDir  string       // eg Default
	theDirectory   string       // eg /home/gilbert/.cj/Default (no trailing path separator)
	theNote        *note.Note   // the current note
	theFound       []*note.Note // the list of found notes
	debugMode      bool
)

type ui struct {
	mainWindow  fyne.Window // Window is an interface
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

func (u *ui) displayText() {
	if len(theNote.Text) == 0 {
		theNote.Load()
	}
	u.noteEntry.SetText(theNote.Text)
}

func (u *ui) setCurrentNote(n *note.Note) {
	theNote = n
	u.displayText()
	u.calendar.Objects[0] = fynex.NewCalendar(theNote.Date, calendarTapped, calendarIsDateImportant)
	u.mainWindow.SetTitle(appTitle(n.Date))
}

func calendarTapped(t time.Time) {
	theNote.SaveIfDirty(theUI.noteEntry.Text)
	theUI.setCurrentNote(note.NewNote(theDirectory, t))
	theUI.foundList.UnselectAll()
	theUI.mainWindow.Canvas().Focus(theUI.noteEntry)
}

func calendarIsDateImportant(t time.Time) bool {
	return t.Year() == theNote.Date.Year() &&
		t.Month() == theNote.Date.Month() &&
		t.Day() == theNote.Date.Day()
}

func find(query string) []*note.Note {
	var found []*note.Note

	if query == "" {
		return found
	}

	// could use xargs to run several directories in parallel?
	// expected output is a list of pathnames, one per line, eg
	// /home/gilbert/.cj/Default/2023/07/04.txt
	// /home/gilbert/.cj/Default/2023/06/18.txt
	// could use ripgrep which is faster
	cmd := exec.Command("grep", // becomes /usr/bin/grep
		"--fixed-strings", // interpret PATTERNS as fixed strings, not regular expressions.
		"--recursive",
		"--ignore-case",
		"--files-with-matches", // print the name of each input file from which output would normally have been printed.
		// regexp.QuoteMeta(query), // don't quote meta if using --fixed-strings
		"-I",                 // don't process binary files
		"--exclude-dir='.*'", // exclude hidden directories
		query,                // don't surround with quotes or double quotes
		theDirectory)
	// fmt.Println(cmd.Path)
	// fmt.Println(cmd.Args)
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
		pathname := stdin.Text()
		// fmt.Println(pathname)
		if strings.HasPrefix(pathname, ".") {
			continue
		}
		n := note.NewNote(theDirectory, pathname)
		found = append(found, n)
	}
	// fmt.Println("-----")
	cmd.Wait() // ignore error return because we're done

	sort.Slice(found, func(i, j int) bool {
		return found[i].Date.Before(found[j].Date)
	})

	return found
}

func daysBetween(m *note.Note, n *note.Note) int {
	return int(n.Date.Sub(m.Date).Hours() / 24)
}

func contains(lst []*note.Note, b *note.Note) bool {
	for _, a := range lst {
		if daysBetween(a, b) == 0 {
			return true
		}
	}
	return false
}

func (u *ui) postFind() {
	if len(theFound) > 0 {
		u.foundList.Select(0)
		u.setCurrentNote(theFound[0])
	} else {
		u.foundList.UnselectAll()
	}
	u.foundList.Refresh()
}

func (u *ui) findEx() {
	var pu *widget.PopUp
	var bfind, bwiden, bnarrow, bexclude, bcancel *widget.Button

	ent := widget.NewEntry()
	ent.PlaceHolder = "Search"
	if len(theFound) == 0 {
		bfind = widget.NewButton("Find", func() {
			results := find(ent.Text)
			if len(results) == 0 {
				return
			}
			var newFound []*note.Note = theFound
			for _, n := range results {
				if !contains(theFound, n) {
					newFound = append(newFound, n)
				}
			}
			theFound = newFound
			u.postFind()
			pu.Hide()
		})
	} else {
		bwiden = widget.NewButton("Widen", func() {
			results := find(ent.Text)
			if len(results) == 0 {
				return
			}
			var newFound []*note.Note = theFound
			for _, n := range results {
				if !contains(theFound, n) {
					newFound = append(newFound, n)
				}
			}
			theFound = newFound
			u.postFind()
			pu.Hide()
		})
		bnarrow = widget.NewButton("Narrow", func() {
			results := find(ent.Text)
			if len(results) == 0 {
				return
			}
			var newFound []*note.Note
			for _, n := range results {
				if contains(theFound, n) {
					newFound = append(newFound, n)
				}
			}
			theFound = newFound
			u.postFind()
			pu.Hide()
		})
		bexclude = widget.NewButton("Exclude", func() {
			results := find(ent.Text)
			if len(results) == 0 {
				return
			}
			var newFound []*note.Note
			for _, n := range results {
				if !contains(theFound, n) {
					newFound = append(newFound, n)
				}
			}
			theFound = newFound
			u.postFind()
			pu.Hide()
		})
	}
	bcancel = widget.NewButton("Cancel", func() {
		pu.Hide()
	})
	var buttons *fyne.Container
	if bfind != nil {
		buttons = container.New(layout.NewHBoxLayout(), bfind, bcancel)
	} else {
		buttons = container.New(layout.NewHBoxLayout(), bwiden, bnarrow, bexclude, bcancel)
	}
	content := container.New(layout.NewVBoxLayout(), ent, buttons)
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
		widget.NewToolbarAction(u.theme.Icon("link"), func() {
			if str := theUI.noteEntry.SelectedText(); str != "" {
				if err := link(str); err != nil {
					dialog.ShowError(err, theUI.mainWindow)
				}
			}
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			t := theNote.Date
			t = t.Add(-time.Hour * 24)
			calendarTapped(t)
		}),
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			calendarTapped(time.Now())
		}),
		widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
			t := theNote.Date
			t = t.Add(time.Hour * 24)
			calendarTapped(t)
		}),
	)

	u.calendar = container.New(layout.NewCenterLayout(), fynex.NewCalendar(theNote.Date, calendarTapped, calendarIsDateImportant))

	u.searchEntry = widget.NewEntry()
	u.searchEntry.PlaceHolder = "Search"
	u.searchEntry.OnChanged = func(str string) {
		theFound = []*note.Note{}
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
		theFound = []*note.Note{}
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
			if theFound[id].Date.Year() == 1 {
				// "no date" shows as Mon 1 Jan 0001
				obj.(*widget.Label).SetText(theFound[id].Pathname)
			} else {
				obj.(*widget.Label).SetText(theFound[id].Date.Format("Mon 2 Jan 2006"))
			}
		},
	)
	u.foundList.OnSelected = func(id widget.ListItemID) {
		theNote.SaveIfDirty(theUI.noteEntry.Text)
		theUI.setCurrentNote(theFound[id])
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
		theDirectory = path.Join(theUserHomeDir, theDataDir, theJournalDir)
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
			theDirectory = path.Join(theUserHomeDir, theDataDir, theJournalDir)
			calendarTapped(time.Now())
			theFound = []*note.Note{}
			u.foundList.Refresh()
		}
	})
}

func (u *ui) searchForHashTags() {
	cmd := exec.Command("grep",
		"--extended-regexp", // because we are using character class
		"--recursive",
		"--ignore-case",
		"--only-matching",    // print only matching parts
		"--no-filename",      // do not prefix output with file name
		"-I",                 // don't process binary files
		"--exclude-dir='.*'", // exclude hidden directories
		"#[[:alnum:]]+",
		theDirectory)
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

	theDirectory = path.Join(theUserHomeDir, theDataDir, theJournalDir)
	// if debugMode {
	// if str, err := os.Executable(); err != nil {
	// 	log.Printf("err: %T, %v\n", err, err)
	// } else {
	// 	log.Printf("str: %T, %v\n", str, str)
	// }
	// log.Println("\nhome:", theUserHomeDir, "\ndata:", theDataDir, "\njournal:", theJournalDir)
	// log.Println(theDirectory)
	// }

	a := app.NewWithID("oddstream.cj")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "today.png",
		StaticContent: todayIconBytes,
	})

	// theTheme := fynex.NewNoteTheme()
	// a.Settings().SetTheme(&theTheme)

	theUI = &ui{mainWindow: a.NewWindow(appTitle(time.Now())), theme: fynex.NewNoteTheme()}
	a.Settings().SetTheme(theUI.theme)
	theNote = note.NewNote(theDirectory, time.Now())

	// shortcuts get swallowed if focus is in the note multiline entry widget
	// don't need this: just tap the 'today' icon in the taskbar
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	theUI.mainWindow.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		theNote.SaveIfDirty(theUI.noteEntry.Text)
	})

	theUI.mainWindow.SetContent(buildUI(theUI))
	theUI.mainWindow.Canvas().Focus(theUI.noteEntry)
	theUI.displayText()

	theUI.mainWindow.Resize(fyne.NewSize(float32(windowWidth), float32(windowHeight)))
	theUI.mainWindow.CenterOnScreen()
	theUI.mainWindow.ShowAndRun()

	// we *do* come here when app quits because window close [x] button pressed
	theNote.SaveIfDirty(theUI.noteEntry.Text)
}
