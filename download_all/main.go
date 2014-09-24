package main

/*
This is an example of how to use Simplenote API to download all
notes.

It downloads all your notes and prints them to stdout.
*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kjk/simplenote"
)

var (
	toJson   bool
	maxNotes = -1 // for debugging, 0 or less means "no limit"
)

func usage() {
	fmt.Printf("usage: download_all -to-json username password\n")
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

func parseFlags() {
	flag.BoolVar(&toJson, "to-json", false, "if true, print result as json")
	flag.Parse()
}

type Notes struct {
	NoteInfos []*simplenote.NoteInfo
	Notes     []*simplenote.Note
}

func main() {
	var api *simplenote.Api
	parseFlags()
	args := flag.Args()
	if true {
		if len(args) != 2 {
			usage()
			return
		}
		api = simplenote.New(args[0], args[1])
	} else {
		api = simplenote.New("foo@bar.com", "password")
	}
	noteInfos, err := api.GetNoteList()
	if err != nil {
		log.Fatalf("api.GetNoteList() returned %q", err)
	}
	if toJson {
		var notes []*simplenote.Note
		for _, ni := range noteInfos {
			note, err := api.GetNoteLatestVersion(ni.Key)
			if err != nil {
				log.Fatalf("api.GetNoteLatestVersion(%q) failed with %q", ni.Key, err)
			}
			notes = append(notes, note)
			maxNotes--
			if maxNotes == 0 {
				break
			}
		}
		v := &Notes{
			NoteInfos: noteInfos,
			Notes:     notes,
		}
		jsonStr, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			log.Fatalf("json.MarshalIndent() failed with %q\n", err)
		}
		fmt.Println(string(jsonStr))
	} else {
		fmt.Printf("You have %d notes\n", len(noteInfos))
		for _, ni := range noteInfos {
			note, err := api.GetNoteLatestVersion(ni.Key)
			if err != nil {
				log.Fatalf("api.GetNoteLatestVersion(%q) failed with %q", ni.Key, err)
			}
			dumpNote(note)
		}
	}
}
