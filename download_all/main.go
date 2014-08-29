package main

/*
This is an example of how to use Simplenote API to download all
notes.

It downloads all your notes and saves them in a single file notes.txt.
*/

import (
	"fmt"
	"log"
	"os"

	"github.com/kjk/simplenote"
)

func usage() {
	fmt.Printf("usage: download_all username password\n")
}

func main() {
	var api *simplenote.Api
	if true {
		if len(os.Args) != 3 {
			usage()
			return
		}
		api = simplenote.New(os.Args[1], os.Args[2])
	} else {
		api = simplenote.New("foo@bar.com", "password")
	}
	notes, err := api.GetNoteListWithLimit(5)
	if err != nil {
		log.Fatalf("api.GetNoteList() returned %q", err)
	} else {
		fmt.Printf("You have %d notes\n", len(notes))
	}
	key := notes[0].Key
	note, err := api.GetNoteLatestVersion(key)
	if err != nil {
		log.Fatalf("api.GetNote(%q) failed with %q", key, err)
	}
	fmt.Printf("\n%s\n", note.Content)
}
