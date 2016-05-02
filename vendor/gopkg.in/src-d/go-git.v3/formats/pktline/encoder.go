package pktline

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrOverflow = errors.New("unexpected string length (overflow)")
)

// Encoder implements a pkt-line format encoder
type Encoder struct {
	lines []string
}

// NewEncoder returns a new Encoder
func NewEncoder() *Encoder {
	return &Encoder{make([]string, 0)}
}

// AddLine encode and adds a line to the encoder
func (e *Encoder) AddLine(line string) error {
	le, err := EncodeFromString(line + "\n")
	if err != nil {
		return err
	}

	e.lines = append(e.lines, le)
	return nil
}

// AddFlush adds a flush-pkt to the encoder
func (e *Encoder) AddFlush() {
	e.lines = append(e.lines, "0000")
}

// Reader returns a string.Reader over the encoder
func (e *Encoder) Reader() *strings.Reader {
	data := strings.Join(e.lines, "")

	return strings.NewReader(data)
}

// EncodeFromString encodes a string to pkt-line format
func EncodeFromString(line string) (string, error) {
	return Encode([]byte(line))
}

// Encode encodes a byte slice to pkt-line format
func Encode(line []byte) (string, error) {
	if line == nil {
		return "0000", nil
	}

	l := len(line) + HeaderLength
	if l > MaxLength {
		return "", ErrOverflow
	}

	return fmt.Sprintf("%04x%s", l, line), nil
}
