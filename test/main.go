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

func joined(tags []string) string {
	return strings.Join(tags, ",")
}

func dumpNote(n *simplenote.Note) {
	fmt.Printf("Key: %s\n", n.Key)
	fmt.Printf("Creation     date: %s\n", n.CreateDate.Format(time.RFC3339))
	fmt.Printf("Modification date: %s\n", n.ModifyDate.Format(time.RFC3339))
	fmt.Printf("Version: %d\n", n.Version)
	if len(n.Tags) > 0 {
		fmt.Printf("Tags: %s\n", joined(n.Tags))
	}
	fmt.Printf("Content: %s\n", strShorten(n.Content, 72))
}

func dumpNoteInfo(ni *simplenote.NoteInfo) {
	fmt.Printf("Key: %s\n", ni.Key)
	fmt.Printf("Creation date: %s\n", ni.CreateDate.Format(time.RFC3339))
	fmt.Printf("Version: %d\n", ni.Version)
	if len(ni.Tags) > 0 {
		fmt.Printf("Tags: %s\n", joined(ni.Tags))
	}
	//n, err := api.GetNoteLatestVersion(ni.Key)
	//fatalIfErr(err)
}

func dumpNotes(notes []*simplenote.NoteInfo) {
	fmt.Printf("have %d notes\n", len(notes))
	for i, ni := range notes {
		if i != 0 {
			fmt.Print("----------------------------\n")
		}
		dumpNoteInfo(ni)
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
	notes, err := api.GetNoteList()
	fatalIfErr(err)
	if len(notes) == 0 {
		return
	}
	dumpNotes(notes)
	for _, ni := range notes {
		err = api.DeleteNote(ni.Key)
		fatalIfErr(err)
	}
	notes, err = api.GetNoteList()
	fatalIfErr(err)
	panicif(len(notes) != 0)
}

func findNoteInfoWithKey(notes []*simplenote.NoteInfo, key string) *simplenote.NoteInfo {
	for _, ni := range notes {
		if ni.Key == key {
			return ni
		}
	}
	return nil
}

func testAll() {
	api = simplenote.New(USER, PWD)
	//deleteAllNotes()
	c := "this is a note"
	c2 := "content 2"
	note, err := api.AddNote(c, nil)
	fatalIfErr(err)
	panicif(note.Content != c)
	notes, err := api.GetNoteList()
	fatalIfErr(err)
	//panicif(len(notes) != 1)
	key := note.Key
	n1 := findNoteInfoWithKey(notes, key)
	panicif(n1 == nil)
	err = api.UpdateContent(key, c2)
	fatalIfErr(err)
	note, err = api.GetNoteLatestVersion(key)
	fatalIfErr(err)
	panicif(note.Content != c2)
	panicif(note.Version != 2)
	tags := []string{"foo", "bar"}
	err = api.UpdateTags(key, tags)
	fatalIfErr(err)
	note, err = api.GetNoteLatestVersion(key)
	fatalIfErr(err)
	panicif(note.Content != c2)
	panicif(note.Version != 3)
	dumpNote(note)
	deleteAllNotes()
}

func testListNotes() {
	api = simplenote.New(USER, PWD)
	for i := 0; i < 20; i++ {
		_, err := api.GetNoteList()
		fatalIfErr(err)
		fmt.Printf(".")
	}
}

func main() {
	fmt.Printf("starting\n")
	testAll()
	fmt.Printf("finished\n")
}
