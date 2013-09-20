package shell

import (
	"bufio"
	"bytes"
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
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a new-line terminated line. Return the line with the newline
		return i + 1, dropCR(data[0 : i+1]), nil
	}

	if atEOF {
		// We have a final, non-terminated line
		return len(data), dropCR(data), nil
	}

	if data[len(data)-1] != '\r' {
		// We have a normal line, just let it tokenize
		return len(data), data, nil
	}

	// We need more data
	return 0, nil, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-2] == '\r' {
		// Trim off the last byte and replace it with a '\n'
		data = data[0 : len(data)-1]
		data[len(data)-1] = '\n'
	}

	return data
}
