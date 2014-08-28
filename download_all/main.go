package main

/*
This is an example of how to use Simplenote API to download all
notes.

It downloads all your notes and saves them in a single file notes.txt.
*/

import (
	"fmt"
	"os"

	"github.com/kjk/simplenote"
)

func usage() {
	fmt.Printf("usage: download_all username password\n")
}

func main() {
	var sn *simplenote.Api
	if true {
		if len(os.Args) != 3 {
			usage()
			return
		}
		sn = simplenote.New(os.Args[1], os.Args[2])
	} else {
		sn = simplenote.New("foo@bar.com", "password")
	}
	notes, err := sn.GetNoteList()
	if err != nil {
		fmt.Printf("sn.GetNoteList() returned %q\n", err)
	} else {
		fmt.Printf("You have %d notes\n", len(notes))
	}
}
