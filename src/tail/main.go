package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hpcloud/tail"
)

var (
	verbose  = flag.Bool("verbose", false, "run in verbose mode")
	filename = flag.String("file", "", "the file to watch")
	timeout  = flag.String("timeout", "30", "the maximum waiting time")
)

func main() {
	flag.Parse()

	if *filename == "" {
		log.Fatal("missing filename: use -file <file>")
	}

	if _, err := strconv.Atoi(*timeout); err == nil {
		*timeout += "s"
	}

	dur, err := time.ParseDuration(*timeout)
	if err != nil {
		log.Fatalf("bad timeout value %q: %#v", *timeout, err)
	}

	exp := os.Args[len(os.Args)-1]
	watchFor, err := regexp.Compile(exp)
	if err != nil {
		log.Fatalf("bad watch regexp %s: %#v", exp, err)
	}

	t, err := tail.TailFile(*filename, tail.Config{
		Follow: true,
		Poll:   true,
	})
	if err != nil {
		log.Fatalf("failed to create log tailer: %#v", err)
	}

	ch := make(chan struct{})

	go func() {
		// Tail the output looking for the expected result
		for line := range t.Lines {
			if line.Err != nil {
				log.Fatalf("log tailer failed: %#v", line.Err)
			}
			if *verbose {
				println(line.Text)
			}
			if watchFor.MatchString(line.Text) {
				if *verbose {
					log.Printf("matched %s OK", exp)
				}
				ch <- struct{}{}
			}
		}
	}()

	select {
	case <-time.After(dur):
		log.Fatalf("Log tailer timed out without seeing any of %s", exp)
	case <-ch:
		os.Exit(0)
	}
}
