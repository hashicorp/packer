package pktline

import (
	"bytes"
	"io/ioutil"
	"strings"

	. "gopkg.in/check.v1"
)

type EncoderSuite struct{}

var _ = Suite(&EncoderSuite{})

func (s *EncoderSuite) TestEncode(c *C) {
	line, err := Encode([]byte("a\n"))
	c.Assert(err, IsNil)
	c.Assert(string(line), Equals, "0006a\n")
}

func (s *EncoderSuite) TestEncodeNil(c *C) {
	line, err := Encode(nil)
	c.Assert(err, IsNil)
	c.Assert(string(line), Equals, "0000")
}

func (s *EncoderSuite) TestEncodeOverflow(c *C) {
	_, err := Encode(bytes.Repeat([]byte{'0'}, MaxLength+1))
	c.Assert(err, Equals, ErrOverflow)
}

func (s *EncoderSuite) TestEncodeFromString(c *C) {
	line, err := EncodeFromString("a\n")
	c.Assert(err, IsNil)
	c.Assert(string(line), Equals, "0006a\n")
}

func (s *EncoderSuite) TestEncoder(c *C) {
	e := NewEncoder()
	c.Assert(e.AddLine("a"), IsNil)
	e.AddFlush()
	c.Assert(e.AddLine("b"), IsNil)

	over := strings.Repeat("0", MaxLength+1)
	c.Assert(e.AddLine(over), Equals, ErrOverflow)

	r := e.Reader()
	a, _ := ioutil.ReadAll(r)
	c.Assert(string(a), Equals, "0006a\n00000006b\n")
}
