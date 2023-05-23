package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type note struct {
	date  time.Time // will be IsZero() for undated notes
	fname string    // the filename this note was loaded from
	title string    // the first line of an undated note
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
		return path.Join(userConfigDir, "oddstream.games", "goldnotebook", "undated", name, ".txt")
	} else {
		return path.Join(userConfigDir, "oddstream.games", "goldnotebook",
			fmt.Sprintf("%04d", n.date.Year()),
			fmt.Sprintf("%02d", n.date.Month()),
			fmt.Sprintf("%02d", n.date.Day()),
			".txt")
	}
}

func NewDateNote(date time.Time) *note {
	return &note{date: date}
}
