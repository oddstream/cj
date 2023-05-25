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
	date time.Time // will be IsZero() for undated notes
	text string    // the text of the note, when loaded
}

// title, which will be the first line of the note (for both dated and undated)
// TODO maybe scan forward incase first line is blank but second line is interesting
// TODO what if several notes have the same first line? overwriting, that's what
func (n *note) title() string {
	title, _, found := strings.Cut(n.text, "\n")
	if !found {
		title = n.text // may be ""
	}
	if title == "" {
		title = "untitled"
	}
	return title
}

func (n *note) fname() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	return path.Join(userHomeDir, NOTEBOOK_DIR,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()),
		fmt.Sprintf("%02d.txt", n.date.Day()))
}

func parseDateFromFname(fname string) time.Time {
	var t = time.Time{}
	lst := strings.Split(fname, string(os.PathSeparator))
	for i := 0; i < len(lst)-3; i++ {
		if lst[i] == NOTEBOOK_DIR {
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

func load(date time.Time) *note {
	n := &note{date: date}
	fname := n.fname()
	file, err := os.Open(fname)
	if err != nil || file == nil {
		log.Print(fname, " does not exist")
	} else {
		fi, err := file.Stat()
		if err != nil {
			log.Fatal(err, " getting FileInfo ", fname)
		}
		if fi.Size() == 0 {
			log.Print(fname, " is empty")
		} else {
			bytes := make([]byte, fi.Size()+8)
			count, err := file.Read(bytes)
			if err != nil {
				log.Fatal(err, " reading ", fname)
			}
			log.Printf("read %d bytes from %s", count, fname)
			n.text = string(bytes[:count])
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err, " closing ", fname)
		}
	}
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

func (n *note) save() {
	fname := n.fname()

	if isStringEmpty(n.text) {
		log.Println("remove", fname)
		os.Remove(fname)
		return
	}

	// make sure the data dir has been created
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := path.Join(userHomeDir, NOTEBOOK_DIR,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()))
	err = os.MkdirAll(dir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	// now save the note text to file
	log.Println("saving", fname)
	file, err := os.Create(fname)
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
}
