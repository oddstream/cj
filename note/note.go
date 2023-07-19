package note

import (
	"log"
	"os"
	"path/filepath"
)

type Note struct {
	Text string // the text of the note, when loaded
}

func (n *Note) Load(pathname string) {
	bytes, _ := os.ReadFile(pathname) // ignore error return because it's ok if pathname does not exist
	n.Text = string(bytes)
}

func (n *Note) Save(pathname string) {
	var err error
	// make sure the data dir has been created
	dir, _ := filepath.Split(pathname)
	// https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err = os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	var file *os.File
	if file, err = os.Create(pathname); err != nil {
		log.Fatal(err)
	}
	if _, err = file.Write([]byte(n.Text)); err != nil {
		log.Fatal(err)
	}
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}
}

func (n *Note) Remove(fname string) {
	os.Remove(fname)
}
