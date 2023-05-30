package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type note struct {
	date time.Time // should never be IsZero()
	text string    // the text of the note, when loaded
}

// title, which will be the first line of the note
func (n *note) title() string {
	var title string
	scanner := bufio.NewScanner(strings.NewReader(n.text))
	for scanner.Scan() {
		title = scanner.Text()
		if len(title) > 0 {
			break
		}
	}
	if title == "" {
		title = "untitled"
	}
	return title
}

// fname of the note, composed of directories and date of note
func (n *note) fname() string {
	return path.Join(theUserHomeDir, theDataDir, theBookDir,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()),
		fmt.Sprintf("%02d.txt", n.date.Day()))
}

// parseDateFromFname
// TODO we can do better than this
func parseDateFromFname(fname string) time.Time {
	var t = time.Time{}
	lst := strings.Split(fname, string(os.PathSeparator))
	for i := 0; i < len(lst)-3; i++ {
		if lst[i] == theBookDir {
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
		if debugMode {
			log.Println(fname, " does not exist")
		}
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
			// if debugMode {
			// 	fmt.Printf("read %d bytes from %s\n", count, fname)
			// }
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
	dir := path.Join(theUserHomeDir, theDataDir, theBookDir,
		fmt.Sprintf("%04d", n.date.Year()),
		fmt.Sprintf("%02d", n.date.Month()))
	err := os.MkdirAll(dir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	// now save the note text to file
	if debugMode {
		log.Println("saving", fname)
	}
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
