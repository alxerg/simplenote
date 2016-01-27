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
	"log"
	"os"

	"github.com/kjk/simplenote"
)

var (
	allVersions = false
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: download_all apiKey username password\n")
}

func parseFlags() {
	flag.BoolVar(&allVersions, "all-versions", false, "if true, download all versions")
	flag.Parse()
}

func printNote(note *simplenote.Note) {
	d, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		log.Fatalf("json.MarshalIndent() failed with '%s'\n", err)
	}
	d = append(d, '\n')
	_, err = os.Stdout.Write(d)
	if err != nil {
		log.Fatalf("os.Stdout.Write() failed with '%s'\n", err)
	}
}

func main() {
	var client *simplenote.Client
	parseFlags()
	args := flag.Args()
	if true {
		if len(args) != 3 {
			usage()
			return
		}
		client = simplenote.NewClient(args[0], args[1], args[2])
	} else {
		client = simplenote.NewClient("api_key", "foo@bar.com", "password")
	}

	notes, err := client.List()
	if err != nil {
		log.Fatalf("c.List() failed with '%s'\n", err)
	}
	for _, note := range notes {
		printNote(note)
		if !allVersions {
			continue
		}
		ver := note.Version - 1
		id := note.ID
		for ver > 0 {
			n, err := client.GetNote(id, ver)
			if err != nil {
				// sometimes older versions don't exist. there doesn't seeme to be
				// a way to list valid versions
				//log.Printf("api.GetNote() failed with '%s'\n", err)
			} else {
				printNote(n)
			}
			ver--
		}
	}
}
