package main

/*
This is an example of how to use Simplenote API to download all
notes.

It downloads all your notes and prints them to stdout.
*/

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kjk/simplenote"
)

func usage() {
	fmt.Printf("usage: download_all username password\n")
}

func joined(tags []string) string {
	return strings.Join(tags, ",")
}

func dumpNote(n *simplenote.Note) {
	fmt.Printf("Key: %s\n", n.Key)
	fmt.Printf("Creation date: %s\n", n.CreateDate.Format(time.RFC3339))
	fmt.Printf("Modification date: %s\n", n.ModifyDate.Format(time.RFC3339))
	fmt.Printf("Version: %d\n", n.Version)
	if len(n.Tags) > 0 {
		fmt.Printf("Tags: %s\n", joined(n.Tags))
	}
	fmt.Printf("Content: %d\n%s\n", len(n.Content), n.Content)
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
	notes, err := api.GetNoteList()
	if err != nil {
		log.Fatalf("api.GetNoteList() returned %q", err)
	}
	fmt.Printf("You have %d notes\n", len(notes))
	for _, ni := range notes {
		note, err := api.GetNoteLatestVersion(ni.Key)
		if err != nil {
			log.Fatalf("api.GetNoteLatestVersion(%q) failed with %q", ni.Key, err)
		}
		dumpNote(note)
	}
}
