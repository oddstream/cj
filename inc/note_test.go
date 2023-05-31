package main

import (
	"testing"
	"time"
)

func TestDateParse(t *testing.T) {
	n := note{
		date: time.Now(),
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	fname := n.fname()
	// fmt.Println(fname)
	tim := parseDateFromFname(fname)
	if tim.Year() != time.Now().Year() {
		t.Error("Not expecting year to be", tim.Year())
	}
	if tim.Month() != time.Now().Month() {
		t.Error("Not expecting month to be", tim.Month())
	}
	if tim.Day() != time.Now().Day() {
		t.Error("Not expecting day to be", tim.Day())
	}
}

func TestNoteFname(t *testing.T) {
	n := note{
		date: time.Date(2023, 05, 23, 0, 0, 0, 0, time.UTC),
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	fname := n.fname()
	if fname != "/home/gilbert/.goldnotebook/2023/05/23.txt" {
		t.Error("Not expecting fname to be", fname)
	}
}

func TestNoteTitle(t *testing.T) {
	n := &note{
		date: time.Now(),
	}
	title := n.title()
	if title != "untitled" {
		t.Error("Not expecting title to be", title)
	}

	n.text = "Shakespeare"
	title = n.title()
	if title != "Shakespeare" {
		t.Error("Not expecting title to be", title)
	}

	n.text = "Now is the winter of our discontent\nMade glorious summer"
	title = n.title()
	if title != "Now is the winter of our discontent" {
		t.Error("Not expecting title to be", title)
	}
}
