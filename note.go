package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	NOTEBOOK_DIR = ".goldnotebook"
)

type note struct {
	date  time.Time // will be IsZero() for undated notes
	fname string    // full path+filename note was loaded from (saved so it can be remove'd if note is empty)
	title string    // the first line of an undated note, as it was when loaded (saved so file can be renamed)
	text  string    // the text of the note, when loaded
}

// getTitle, which will be the first line of the note (for both dated and undated)
// TODO maybe scan forward incase first line is blank but second line is interesting
// TODO what if several notes have the same first line?
func (n *note) getTitle() string {
	title, _, found := strings.Cut(n.text, "\n")
	if !found {
		title = n.text // may be ""
	}
	if title == "" {
		title = "untitled"
	}
	return title
}

func (n *note) getFname() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	if n.date.IsZero() {
		name := n.getTitle()
		if name == "" {
			return ""
		}
		return path.Join(userHomeDir, NOTEBOOK_DIR,
			"undated",
			name+".txt") // TODO could be .md
	} else {
		return path.Join(userHomeDir, NOTEBOOK_DIR,
			"dated",
			fmt.Sprintf("%04d", n.date.Year()),
			fmt.Sprintf("%02d", n.date.Month()),
			fmt.Sprintf("%02d.txt", n.date.Day()))
	}
}

func parseDateFromFname(fname string) time.Time {
	var t = time.Time{}
	lst := strings.Split(fname, string(os.PathSeparator))
	for i := 0; i < len(lst)-3; i++ {
		if lst[i] == "dated" {
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

func load(fname string) *note {
	if fname == "" {
		log.Fatal("cannot load a note with no filename")
	}

	n := &note{}

	file, err := os.Open(fname)
	if err != nil || file == nil {
		log.Print(fname, " does not exist")
		return nil
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err, " getting FileInfo ", fname)
	}
	if fi.Size() == 0 {
		log.Print(fname, " is empty")
	} else {
		bytes := make([]byte, fi.Size()+8)
		_, err = file.Read(bytes)
		if err != nil {
			log.Fatal(err, " reading ", fname)
		}
		n.text = string(bytes)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err, " closing ", fname)
	}
	n.fname = fname
	n.title = n.getTitle()
	n.date = parseDateFromFname(fname)
	return n
}

func loadUndatedNote(title string) *note {
	UserHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	return load(path.Join(UserHomeDir, NOTEBOOK_DIR, "undated", title+".txt"))
}

func loadDatedNote(date time.Time) *note {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	n := load(path.Join(userHomeDir, NOTEBOOK_DIR,
		"dated",
		fmt.Sprintf("%04d", date.Year()),
		fmt.Sprintf("%02d", date.Month()),
		fmt.Sprintf("%02d.txt", date.Day())))
	n.date = date
	return n
}

func isStringEmpty(str string) bool {
	for _, r := range str {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// save note (detect name change if undated) (remove'd file for empty note)
func (n *note) save() {
	if isStringEmpty(n.text) {
		os.Remove(n.fname)
		return
	}
	// make sure the config dir has been created
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	var dir string
	if n.date.IsZero() {
		dir = path.Join(userHomeDir, NOTEBOOK_DIR, "undated")
	} else {
		dir = path.Join(userHomeDir, NOTEBOOK_DIR,
			"dated",
			fmt.Sprintf("%04d", n.date.Year()),
			fmt.Sprintf("%02d", n.date.Month()))
	}
	err = os.MkdirAll(dir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	// now save the note text to file
	file, err := os.Create(n.getFname()) // nb use freshly-generated title, not saved one
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write([]byte(n.text))
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	// delete the old file if the note is undated and title has changed
	if n.date.IsZero() {
		if n.title != n.getTitle() {
			os.Remove(n.fname)
			n.title = n.getTitle()
			n.fname = n.getFname()
		}
	}
}

func NewDateNote(date time.Time) *note {
	return &note{date: date}
}
