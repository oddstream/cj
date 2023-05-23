package main

import (
	"testing"
	"time"
)

func TestNoteFname(t *testing.T) {
	n := &note{
		date: time.Now(),
		text: "Now is the winter of our discontent\nMade glorious summer",
	}

	fname := n.getFname()
	println(fname)
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
