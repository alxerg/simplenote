package main

/*
This is an example of how to use Simplenote API to download all
notes.

It downloads all your notes and prints them to stdout in json form.
*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kjk/simplenote"
)

var (
	flgAllVersions = false
	flgVerbose     = false
	fileName       = "notes.json"
	previousNotes  map[string]bool
)

func init() {
	previousNotes = make(map[string]bool)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: download_all apiKey username password\n")
}

func key(id string, version int) string {
	return fmt.Sprintf("%s-%d", id, version)
}

func noteKey(n *simplenote.Note) string {
	return key(n.ID, n.Version)
}

func wasImported(n *simplenote.Note) bool {
	return previousNotes[noteKey(n)]
}

func wasImported2(id string, version int) bool {
	return previousNotes[key(id, version)]
}

func loadPreviousNotes() error {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		var n simplenote.Note
		err = dec.Decode(&n)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		previousNotes[noteKey(&n)] = true
	}
}

func parseFlags() {
	flag.BoolVar(&flgAllVersions, "all-versions", false, "if true, download all versions")
	flag.BoolVar(&flgVerbose, "verbose", false, "if true, show debug info")
	flag.Parse()
}

func writeNote(file *os.File, note *simplenote.Note) bool {
	if wasImported(note) {
		return false
	}
	d, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		log.Fatalf("json.MarshalIndent() failed with '%s'\n", err)
	}
	d = append(d, '\n')
	_, err = file.Write(d)
	if err != nil {
		log.Fatalf("file.Write() failed with '%s'\n", err)
	}
	return true
}

type logger struct {
	file *os.File
}

func newLogger(path string) *logger {
	var err error
	l := &logger{}
	l.file, err = os.Create(path)
	if err != nil {
		return nil
	}
	return l
}

func (l *logger) Log(s string) {
	if l != nil && l.file != nil {
		l.file.WriteString(s)
	}
}

func (l *logger) Close() {
	if l != nil && l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

func main() {
	var client *simplenote.Client
	parseFlags()
	args := flag.Args()
	if len(args) != 3 {
		usage()
		return
	}
	loadPreviousNotes()
	fmt.Printf("all versions: %v. Previously loaded: %d\n", flgAllVersions, len(previousNotes))
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("os.OpenFile('%s') failed with '%s'\n", fileName, err)
	}
	defer file.Close()
	client = simplenote.NewClient(args[0], args[1], args[2])
	lgr := newLogger("log.txt")
	defer lgr.Close()
	if lgr != nil {
		client.Logger = lgr
	}

	notes, err := client.List()
	if err != nil {
		log.Fatalf("c.List() failed with '%s'\n", err)
	}
	nNotes := 0
	nNotesTotal := 0
	nVersions := 0
	nVersionsTotal := 0
	for _, note := range notes {
		id := note.ID

		if !flgAllVersions {
			for ver := 1; ver < note.Version; ver++ {
				if wasImported2(id, ver) {
					nVersionsTotal++
					continue
				}
				n, err := client.GetNote(id, ver)
				if err != nil {
					// sometimes older versions don't exist. there doesn't seeme to be
					// a way to list valid versions
					//log.Printf("api.GetNote() failed with '%s'\n", err)
					continue
				}
				nVersionsTotal++
				nVersions++
				didWrite := writeNote(file, n)
				if !didWrite {
					log.Fatalf("unexpectedly didWrite on note %v\n", n)
				}
			}
		}
		didWrite := writeNote(file, note)
		nNotesTotal++
		if didWrite {
			nNotes++
		}
	}
	if flgAllVersions {
		fmt.Printf("Imported %d new notes and %d new versions, %d total notes, %d total versions\n", nNotes, nVersions, nNotesTotal, nVersionsTotal)

	} else {
		fmt.Printf("Imported %d new notes, %d total notes\n", nNotes, nNotesTotal)
	}
}
