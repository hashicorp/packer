package packfile

import (
	"bufio"
	"fmt"
	"io"
)

type trackingReader struct {
	r        io.Reader
	position int64
}

func NewTrackingReader(r io.Reader) *trackingReader {
	return &trackingReader{
		r: bufio.NewReader(r),
	}
}

func (t *trackingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if err != nil {
		return 0, err
	}

	t.position += int64(n)
	return n, err
}

func (t *trackingReader) ReadByte() (c byte, err error) {
	var p [1]byte
	n, err := t.r.Read(p[:])
	if err != nil {
		return 0, err
	}

	if n > 1 {
		return 0, fmt.Errorf("read %d bytes, should have read just 1", n)
	}

	t.position++
	return p[0], nil
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
