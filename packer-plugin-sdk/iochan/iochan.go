package iochan

import (
	"bufio"
	"io"
)

// LineReader takes an io.Reader and produces the contents of the reader on the
// returned channel. Internally bufio.NewScanner is used, io.ScanLines parses
// lines and returns them without carriage return. Scan can panic if the split
// function returns too many empty tokens without advancing the input.
//
// The channel will be closed either by reaching the end of the input or an
// error.
func LineReader(r io.Reader) <-chan string {
	ch := make(chan string)

	go func() {
		scanner := bufio.NewScanner(r)
		defer close(ch)

		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()

	return ch
}
