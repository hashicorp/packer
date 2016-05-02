package pktline

import (
	"errors"
	"io"
	"strconv"
)

var (
	ErrUnderflow     = errors.New("unexpected string length (underflow)")
	ErrInvalidHeader = errors.New("invalid header")
	ErrInvalidLen    = errors.New("invalid length")
)

// Decoder implements a pkt-line format decoder
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new Decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// ReadLine reads and return one pkt-line line from the reader
func (d *Decoder) ReadLine() (string, error) {
	return d.readLine()
}

func (d *Decoder) readLine() (string, error) {
	raw := make([]byte, HeaderLength)
	if _, err := d.r.Read(raw); err != nil {
		return "", err
	}

	header, err := strconv.ParseInt(string(raw), 16, 16)
	if err != nil {
		return "", ErrInvalidHeader
	}

	if header == 0 {
		return "", nil
	}

	exp := int(header - HeaderLength)
	if exp < 0 {
		return "", ErrInvalidLen
	}

	line := make([]byte, exp)
	if read, err := d.r.Read(line); err != nil {
		return "", err
	} else if read != exp {
		return "", ErrUnderflow
	}

	return string(line), nil
}

// ReadBlock reads and return multiple pkt-line lines, it stops at the end
// of the reader or if a flush-pkt is reached
func (d *Decoder) ReadBlock() ([]string, error) {
	var o []string

	for {
		line, err := d.readLine()
		if err == io.EOF {
			return o, nil
		}

		if err != nil {
			return o, err
		}

		if err == nil && line == "" {
			return o, nil
		}

		o = append(o, line)
	}
}

// ReadAll read and returns all the lines
func (d *Decoder) ReadAll() ([]string, error) {
	result, err := d.ReadBlock()
	if err != nil {
		return result, err
	}

	for {
		lines, err := d.ReadBlock()
		if err == io.EOF {
			return result, nil
		}

		if err != nil {
			return result, err
		}

		if err == nil && len(lines) == 0 {
			return result, nil
		}

		result = append(result, lines...)
	}
}
