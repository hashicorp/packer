package git

import (
	"io"
	"strings"
)

// countLines returns the number of lines in a string à la git, this is
// The newline character is assumed to be '\n'.  The empty string
// contains 0 lines.  If the last line of the string doesn't end with a
// newline, it will still be considered a line.
func countLines(s string) int {
	if s == "" {
		return 0
	}

	nEOL := strings.Count(s, "\n")
	if strings.HasSuffix(s, "\n") {
		return nEOL
	}

	return nEOL + 1
}

// checkClose is used with defer to close the given io.Closer and check its
// returned error value. If Close returns an error and the given *error
// is not nil, *error is set to the error returned by Close.
//
// checkClose is typically used with named return values like so:
//
//   func do(obj *Object) (err error) {
//     w, err := obj.Writer()
//     if err != nil {
//       return nil
//     }
//     defer checkClose(w, &err)
//     // work with w
//   }
func checkClose(c io.Closer, err *error) {
	if cerr := c.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}
