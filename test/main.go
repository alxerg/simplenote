package main

/*
Test program to exercise the APIs
*/

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kjk/simplenote"
)

const (
	USER = "simplenote-test@lordofhosts.de"
	PWD  = "foobar"
)

var (
	api *simplenote.Api
)

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func strShorten(s string, max int) string {
	if len(s) < max {
		return s
	}
	return s[:max-3] + "..."
}

func dumpNotes() {
	notes, err := api.GetNoteListWithLimit(5)
	fatalIfErr(err)
	fmt.Printf("have %d notes\n", len(notes))
	for i, ni := range notes {
		if i != 0 {
			fmt.Print("----------------------------\n")
		}
		fmt.Printf("Key: %s\n", ni.Key)
		fmt.Printf("Creation date: %s\n", ni.CreateDate.Format(time.RFC1123))
		if len(ni.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(ni.Tags, ","))
		}
		n, err := api.GetNoteLatestVersion(ni.Key)
		fatalIfErr(err)
		fmt.Printf("Content: %s\n", strShorten(n.Content, 72))
	}
}

func deleteAllNotes() {
	dumpNotes()
	notes, err := api.GetNoteListWithLimit(5)
	fatalIfErr(err)
	for _, ni := range notes {
		err = api.DeleteNote(ni.Key)
		fatalIfErr(err)
	}
}

func main() {
	fmt.Printf("starting\n")
	api = simplenote.New(USER, PWD)
	deleteAllNotes()
	note, err := api.AddNote("this is a note", nil)
	fatalIfErr(err)
	fmt.Printf("%#v\n", note)
	fmt.Printf("finished\n")
}
