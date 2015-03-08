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
	maxNotes = -1 // for debugging, 0 or less means "no limit"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: download_all apiKey username password\n")
}

func parseFlags() {
	//flag.BoolVar(&toJson, "to-json", false, "if true, print result as json")
	flag.Parse()
}

func main() {
	var c *simplenote.Client
	parseFlags()
	args := flag.Args()
	if true {
		if len(args) != 3 {
			usage()
			return
		}
		c = simplenote.NewClient(args[0], args[1], args[2])
	} else {
		c = simplenote.NewClient("api_key", "foo@bar.com", "password")
	}

	notes, err := c.List()
	if err != nil {
		log.Fatalf("c.List() failed with '%s'\n", err)
	}
	for _, n := range notes {
		/*n, err := api.GetNote(note.ID, note.V)
		if err != nil {
			log.Fatalf("api.GetNote() failed with '%s'\n", err)
		}*/
		d, err := json.MarshalIndent(n, "", "  ")
		if err != nil {
			log.Fatalf("json.MarshalIndent() failed with '%s'\n", err)
		}
		d = append(d, '\n')
		_, err = os.Stdout.Write(d)
		if err != nil {
			log.Fatalf("os.Stdout.Write() failed with '%s'\n", err)
		}
	}
}
