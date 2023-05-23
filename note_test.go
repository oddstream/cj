package main

import (
	"testing"
	"time"
)

func TestNoteFname(t *testing.T) {
	n := note{
		text: "Now is the winter of our discontent\nMade glorious summer",
	}
	fname := n.getFname()
	if fname != "/home/gilbert/.config/oddstream.games/goldnotebook/undated/Now is the winter of our discontent.txt" {
		t.Error("Not expecting fname to be", fname)
	}

	n.date = time.Date(2023, 05, 23, 1, 1, 1, 1, time.UTC)
	fname = n.getFname()
	if fname != "/home/gilbert/.config/oddstream.games/goldnotebook/2023/05/23.txt" {
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

	n.text = ""
	n.title = ""

	n.loadUndated("Now is the winter of our discontent")
	if n.title != "Now is the winter of our discontent" {
		t.Error("Not expecting title to be", n.title)
	}
}
