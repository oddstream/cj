package main

import (
	"fmt"
	"testing"
	"time"
)

func TestDateParse(t *testing.T) {
	n := note{
		date: time.Now(),
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	fname := n.getFname()
	fmt.Println(fname)
	tim := parseDateFromFname(fname)
	fmt.Println(tim)
}

func TestNoteFname(t *testing.T) {
	n := note{
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	fname := n.getFname()
	if fname != "/home/gilbert/.goldnotebook/undated/Now is the winter of our discontent.txt" {
		t.Error("Not expecting fname to be", fname)
	}

	n.date = time.Date(2023, 05, 23, 1, 1, 1, 1, time.UTC)
	fname = n.getFname()
	if fname != "/home/gilbert/.goldnotebook/2023/05/23.txt" {
		t.Error("Not expecting fname to be", fname)
	}
}

func TestNoteTitle(t *testing.T) {
	n := &note{
		date: time.Now(),
	}
	title := n.getTitle()
	if title != "" {
		t.Error("Not expecting title to be", title)
	}

	n.text = "Shakespeare"
	title = n.getTitle()
	if title != "Shakespeare" {
		t.Error("Not expecting title to be", title)
	}

	n.text = "Now is the winter of our discontent\nMade glorious summer"
	title = n.getTitle()
	if title != "Now is the winter of our discontent" {
		t.Error("Not expecting title to be", title)
	}
}

func TestNoteSave(t *testing.T) {
	n := &note{
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	n.title = n.getTitle()
	n.fname = n.getFname()
	n.save()

	n = &note{
		date: time.Now(),
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	n.title = n.getTitle()
	n.fname = n.getFname()
	n.save()
}

func TestNoteLoad(t *testing.T) {
	n := &note{
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	n.title = n.getTitle()
	n.fname = n.getFname()
	n.save()

	n = loadUndatedNote("Now is the winter of our discontent")
	if n.title != "Now is the winter of our discontent" {
		t.Error("Not expecting title to be", n.title)
	}
}
