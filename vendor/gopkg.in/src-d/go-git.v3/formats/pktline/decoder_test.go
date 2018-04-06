package pktline

import (
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type DecoderSuite struct{}

var _ = Suite(&DecoderSuite{})

func (s *DecoderSuite) TestReadLine(c *C) {
	j := NewDecoder(strings.NewReader("0006a\n"))

	line, err := j.ReadLine()
	c.Assert(err, IsNil)
	c.Assert(line, Equals, "a\n")
}

func (s *DecoderSuite) TestReadLineInvalidHeader(c *C) {
	j := NewDecoder(strings.NewReader("foo\n"))

	_, err := j.ReadLine()
	c.Assert(err, Equals, ErrInvalidHeader)
}

func (s *DecoderSuite) TestReadLineBufferUnderflow(c *C) {
	j := NewDecoder(strings.NewReader("00e7a\n"))

	line, err := j.ReadLine()
	c.Assert(err, Equals, ErrUnderflow)
	c.Assert(line, Equals, "")
}

func (s *DecoderSuite) TestReadLineInvalidLen(c *C) {
	j := NewDecoder(strings.NewReader("0001foo\n"))

	line, err := j.ReadLine()
	c.Assert(err, Equals, ErrInvalidLen)
	c.Assert(line, Equals, "")
}

func (s *DecoderSuite) TestReadBlock(c *C) {
	j := NewDecoder(strings.NewReader("0006a\n"))

	lines, err := j.ReadBlock()
	c.Assert(err, IsNil)
	c.Assert(lines, HasLen, 1)
	c.Assert(lines[0], Equals, "a\n")
}

func (s *DecoderSuite) TestReadBlockWithFlush(c *C) {
	j := NewDecoder(strings.NewReader("0006a\n0006b\n00000006c\n"))

	lines, err := j.ReadBlock()
	c.Assert(err, IsNil)
	c.Assert(lines, HasLen, 2)
	c.Assert(lines[0], Equals, "a\n")
	c.Assert(lines[1], Equals, "b\n")
}

func (s *DecoderSuite) TestReadBlockInvalidLen(c *C) {
	j := NewDecoder(strings.NewReader("0001foo\n"))

	lines, err := j.ReadBlock()
	c.Assert(err, Equals, ErrInvalidLen)
	c.Assert(lines, HasLen, 0)
}

func (s *DecoderSuite) TestReadAll(c *C) {
	j := NewDecoder(strings.NewReader("0006a\n0006b\n00000006c\n0006d\n0006e\n"))

	lines, err := j.ReadAll()
	c.Assert(err, IsNil)
	c.Assert(lines, HasLen, 5)
	c.Assert(lines[0], Equals, "a\n")
	c.Assert(lines[1], Equals, "b\n")
	c.Assert(lines[2], Equals, "c\n")
	c.Assert(lines[3], Equals, "d\n")
	c.Assert(lines[4], Equals, "e\n")
}
