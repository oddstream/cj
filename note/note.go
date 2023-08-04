package note

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"oddstream.cj/util"
)

type Note struct {
	Text     string
	Pathname string
	Date     time.Time
}

func NewNote(home, data, journal string, obj any) *Note {
	n := &Note{}
	switch v := obj.(type) {
	case string:
		n.Pathname = v
		prefix := path.Join(home, data, journal) + "/"
		str := strings.TrimPrefix(v, prefix)
		ext := filepath.Ext(str) // includes .
		str = strings.TrimSuffix(str, ext)
		lst := strings.Split(str, string(os.PathSeparator))
		if len(lst) == 3 {
			y, _ := strconv.Atoi(lst[0])
			m, _ := strconv.Atoi(lst[1])
			d, _ := strconv.Atoi(lst[2])
			n.Date = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
		}
	case time.Time:
		n.Date = v
		n.Pathname = path.Join(home, data, journal,
			fmt.Sprintf("%04d", v.Year()),
			fmt.Sprintf("%02d", v.Month()),
			fmt.Sprintf("%02d.txt", v.Day()))
	}
	return n
}

func (n *Note) Load() {
	bytes, _ := os.ReadFile(n.Pathname) // ignore error return because it's ok if pathname does not exist
	n.Text = string(bytes)
}

func (n *Note) Save() {
	var err error
	// make sure the data dir has been created
	dir, _ := filepath.Split(n.Pathname)
	// https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	if err = os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	// if path is already a directory, MkdirAll does nothing and returns nil

	var file *os.File
	if file, err = os.Create(n.Pathname); err != nil {
		log.Fatal(err)
	}
	if _, err = file.Write([]byte(n.Text)); err != nil {
		log.Fatal(err)
	}
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}
}

func (n *Note) SaveIfDirty(newText string) {
	if newText != n.Text {
		if util.IsStringEmpty(newText) {
			n.Remove()
		} else {
			n.Text = newText
			n.Save()
		}
	}
}

func (n *Note) Remove() {
	os.Remove(n.Pathname)
}
