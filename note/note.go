package note

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Note struct {
	Text string // the text of the note, when loaded
}

// Title, which will be the first line of the note
func (n *Note) Title() string {
	var title string
	scanner := bufio.NewScanner(strings.NewReader(n.Text))
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

func (n *Note) Load(fname string) {
	file, err := os.Open(fname)
	if err != nil || file == nil {
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
			n.Text = string(bytes[:count])
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err, " closing ", fname)
		}
	}
}

func isStringEmpty(str string) bool {
	for _, r := range str {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func (n *Note) Save(fname string) {

	if isStringEmpty(n.Text) {
		log.Println("remove", fname)
		os.Remove(fname)
		return
	}

	// make sure the data dir has been created
	dir, filename := filepath.Split(fname)
	err := os.MkdirAll(dir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	// now save the note text to file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write([]byte(n.Text))
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}
