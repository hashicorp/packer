package objfile

import (
	"errors"
	"io"
	"strconv"

	"gopkg.in/src-d/go-git.v3/core"
)

var (
	// ErrClosed is returned when the objfile Reader or Writer is already closed.
	ErrClosed = errors.New("objfile: already closed")
	// ErrHeader is returned when the objfile has an invalid header.
	ErrHeader = errors.New("objfile: invalid header")
	// ErrNegativeSize is returned when a negative object size is declared.
	ErrNegativeSize = errors.New("objfile: negative object size")
)

type header struct {
	t    core.ObjectType
	size int64
}

func (h *header) Read(r io.Reader) error {
	t, err := h.readSlice(r, ' ')
	if err != nil {
		return err
	}

	h.t, err = core.ParseObjectType(string(t))
	if err != nil {
		return err
	}

	size, err := h.readSlice(r, 0)
	if err != nil {
		return err
	}

	h.size, err = strconv.ParseInt(string(size), 10, 64)
	if err != nil {
		return ErrHeader
	}

	if h.size < 0 {
		return ErrNegativeSize
	}

	return nil
}

func (h *header) Write(w io.Writer) error {
	b := h.t.Bytes()
	b = append(b, ' ')
	b = append(b, []byte(strconv.FormatInt(h.size, 10))...)
	b = append(b, 0)
	_, err := w.Write(b)
	return err
}

// readSlice reads one byte at a time from r until it encounters delim or an
// error.
func (h *header) readSlice(r io.Reader, delim byte) ([]byte, error) {
	var buf [1]byte
	value := make([]byte, 0, 16)
	for {
		if n, err := r.Read(buf[:]); err != nil && (err != io.EOF || n == 0) {
			if err == io.EOF {
				return nil, ErrHeader
			}
			return nil, err
		}
		if buf[0] == delim {
			return value, nil
		}
		value = append(value, buf[0])
	}
}
