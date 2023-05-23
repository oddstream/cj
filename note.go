package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
	"unicode"
)

type note struct {
	date  time.Time // will be IsZero() for undated notes
	fname string    // full path+filename note was loaded from (saved so it can be remove'd if note is empty)
	title string    // the first line of an undated note, as it was when loaded (saved so file can be renamed)
	text  string    // the text of the note, when loaded
}

// getTitle, which will be the first line of the note
func (n *note) getTitle() string {
	title, _, found := strings.Cut(n.text, "\n")
	if !found {
		title = n.text // may be ""
	}
	return title
}

func (n *note) getFname() string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Panic(err)
	}
	if n.date.IsZero() {
		name := n.getTitle()
		if name == "" {
			return ""
		}
		return path.Join(userConfigDir, "oddstream.games", "goldnotebook", "undated", name+".txt")
	} else {
		return path.Join(userConfigDir, "oddstream.games", "goldnotebook",
			fmt.Sprintf("%04d", n.date.Year()),
			fmt.Sprintf("%02d", n.date.Month()),
			fmt.Sprintf("%02d.txt", n.date.Day()))
	}
}

func (n *note) load() {
	file, err := os.Open(n.fname)
	if err != nil || file == nil {
		log.Print(n.fname, " does not exist")
		return
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err, " getting FileInfo ", n.fname)
	}
	if fi.Size() == 0 {
		log.Print(n.fname, " is empty")
	} else {
		bytes := make([]byte, fi.Size()+8)
		_, err = file.Read(bytes)
		if err != nil {
			log.Fatal(err, " reading ", n.fname)
		}
		n.text = string(bytes)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err, " closing ", n.fname)
	}
	n.title = n.getTitle()
}

func (n *note) loadUndated(title string) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Panic(err)
	}
	n.fname = path.Join(userConfigDir, "oddstream.games", "goldnotebook", "undated", title+".txt")
	n.load()
}

func (n *note) loadDated(date time.Time) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Panic(err)
	}
	n.fname = path.Join(userConfigDir, "oddstream.games", "goldnotebook",
		fmt.Sprintf("%04d", date.Year()),
		fmt.Sprintf("%02d", date.Month()),
		fmt.Sprintf("%02d.txt", date.Day()))
	n.load()
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
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	var dir string
	if n.date.IsZero() {
		dir = path.Join(userConfigDir, "oddstream.games", "goldnotebook", "undated")
	} else {
		dir = path.Join(userConfigDir, "oddstream.games", "goldnotebook",
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
