package note

import (
	"log"
	"os"
	"path/filepath"
	"unicode"
)

type Note struct {
	Text string // the text of the note, when loaded
}

func (n *Note) Load(fname string) {
	// println("load", fname)
	file, err := os.Open(fname)
	if err != nil || file == nil {
		log.Println(err)
	} else {
		fi, err := file.Stat()
		if err != nil {
			log.Fatal(err, " getting FileInfo ", fname)
		}
		if fi.Size() == 0 {
			log.Println(fname, " is empty")
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
	// println("save", fname)
	if isStringEmpty(n.Text) {
		n.Remove(fname)
		return
	}

	// make sure the data dir has been created
	dir, _ := filepath.Split(fname)
	err := os.MkdirAll(dir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	// now save the note text to file
	file, err := os.Create(fname)
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

func (n *Note) Remove(fname string) {
	println("remove", fname)
	os.Remove(fname)
}
