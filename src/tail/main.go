package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hpcloud/tail"
)

var verbose = flag.Bool("verbose", false, "run in verbose mode")

func main() {
	flag.Parse()

	minArgs := 3
	if yes, _ := regexp.MatchString("verbose", os.Args[1]); yes {
		minArgs++
	}

	switch {
	case len(os.Args) < minArgs-1:
		log.Fatal("Missing log filename")
	case len(os.Args) < minArgs:
		log.Fatal("No log entries to finish on")
	}

	// Generate regex to finish on
	cases := make([]*regexp.Regexp, len(os.Args)-(minArgs-1))
	for i, c := range os.Args[minArgs-1:] {
		cases[i] = regexp.MustCompile(strings.Trim(c, "\""))
	}

	t, err := tail.TailFile(os.Args[minArgs-2], tail.Config{
		Follow: true,
		Poll:   true,
	})
	if err != nil {
		log.Fatalf("log tailer failed: %#v", err)
	}

	ch := make(chan struct{})

	go func() {
		// Tail the output looking for the expected result
		for line := range t.Lines {
			if line.Err != nil {
				log.Fatalf("log tailer failed: %#v", line.Err)
			}
			if *verbose {
				log.Println(line.Text)
			}
			for _, c := range cases {
				if c.MatchString(line.Text) {
					ch <- struct{}{}
					return
				}
			}
		}
	}()

	select {
	case <-time.After(30 * time.Second):
		log.Fatalf("Log tailer timed out without seeing any of %#v",
			os.Args[minArgs-1:],
		)
	case <-ch:
		os.Exit(0)
	}
}
