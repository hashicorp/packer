package shell

import (
	"bufio"
	"io"
	"sync"
)

// UnixReader is a Reader implementation that automatically converts
// Windows line endings to Unix line endings.
type UnixReader struct {
	Reader io.Reader

	buf     []byte
	once    sync.Once
	scanner *bufio.Scanner
}

func (r *UnixReader) Read(p []byte) (n int, err error) {
	// Create the buffered reader once
	r.once.Do(func() {
		r.scanner = bufio.NewScanner(r.Reader)
		r.scanner.Split(scanUnixLine)
	})

	// If we have no data in our buffer, scan to the next token
	if len(r.buf) == 0 {
		if !r.scanner.Scan() {
			err = r.scanner.Err()
			if err == nil {
				err = io.EOF
			}

			return 0, err
		}

		r.buf = r.scanner.Bytes()
	}

	// Write out as much data as we can to the buffer, storing the rest
	// for the next read.
	n = len(p)
	if n > len(r.buf) {
		n = len(r.buf)
	}
	copy(p, r.buf)
	r.buf = r.buf[n:]

	return
}

// scanUnixLine is a bufio.Scanner SplitFunc. It tokenizes on lines, but
// only returns unix-style lines. So even if the line is "one\r\n", the
// token returned will be "one\n".
func scanUnixLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	if advance == 0 {
		// If we reached the end of a line without a newline, then
		// just return as it is. Otherwise the Scanner will keep trying
		// to scan, blocking forever.
		return
	}

	return advance, append(token, '\n'), err
}
