package main

import (
	"flag"
	"fmt"
	"gopkg.in/xmlpath.v2"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var all = flag.Bool("all", false, "print all occurrences rather than the first one")
var trim = flag.Bool("trim", false, "trim spaces around results")
var line = flag.Bool("line", false, "reformat each match as a single line")
var quiet = flag.Bool("q", false, "run quietly with no stdout output")

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "usage: webpath <xpath> <url>\n")
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var whitespace = regexp.MustCompile("[ \t\n]+")

func run() error {
	args := flag.Args()

	path, err := xmlpath.Compile(args[0])
	if err != nil {
		return err
	}

	loc := args[1]

	var body io.Reader
	if strings.HasPrefix(loc, "https:") || strings.HasPrefix(loc, "http:") {
		resp, err := http.Get(args[1])
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body = resp.Body
	} else {
		file, err := os.Open(loc)
		if err != nil {
			return err
		}
		defer file.Close()
		body = file
	}

	n, err := xmlpath.ParseHTML(body)
	if err != nil {
		return err
	}

	iter := path.Iter(n)
	ok := false
	for iter.Next() {
		ok = true
		if *quiet {
			break
		}
		s := iter.Node().String()
		if *line {
			s = strings.TrimSpace(whitespace.ReplaceAllString(s, " "))
		} else if *trim {
			s = strings.TrimSpace(s)
		}
		fmt.Println(s)
		if !*all {
			break
		}
	}
	if !ok {
		os.Exit(1)
	}
	return nil
}
