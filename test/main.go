package main

/*
Test program to exercise the APIs
*/

import (
	"fmt"
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

func panicif(cond bool) {
	if cond {
		panic("error")
	}
}

func fatalIfErr(err error) {
	if err == nil {
		return
	}
	s := err.Error()
	fmt.Println(s)
	panic(s)
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
		fmt.Printf("Creation date: %s\n", ni.CreateDate.Format(time.RFC3339))
		if len(ni.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(ni.Tags, ","))
		}
		n, err := api.GetNoteLatestVersion(ni.Key)
		fatalIfErr(err)
		fmt.Printf("Content: %s\n", strShorten(n.Content, 72))
	}
}

func testTrashNote(key string) {
	n, err := api.TrashNote(key)
	fatalIfErr(err)
	if !n.IsDeleted {
		panic(fmt.Sprint("%#v is not deleted\n", n))
	} else {
		//fmt.Printf("%s has been trashed\n", key)
	}
}

func deleteAllNotes() {
	dumpNotes()
	notes, err := api.GetNoteList()
	fatalIfErr(err)
	if len(notes) == 0 {
		return
	}
	for _, ni := range notes {
		err = api.DeleteNote(ni.Key)
		fatalIfErr(err)
	}
	notes, err = api.GetNoteList()
	fatalIfErr(err)
	panicif(len(notes) != 0)
}

func main() {
	fmt.Printf("starting\n")
	api = simplenote.New(USER, PWD)
	deleteAllNotes()
	c := "this is a note"
	note, err := api.AddNote(c, nil)
	fatalIfErr(err)
	panicif(note.Content != c)
	notes, err := api.GetNoteList()
	fatalIfErr(err)
	panicif(len(notes) != 1)
	n1 := notes[0]
	panicif(note.Key != n1.Key)
	deleteAllNotes()
	fmt.Printf("finished\n")
}
