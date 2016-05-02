package objfile

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/core"
)

type SuiteWriter struct{}

var _ = Suite(&SuiteWriter{})

func (s *SuiteWriter) TestWriteObjfile(c *C) {
	for k, fixture := range objfileFixtures {
		comment := fmt.Sprintf("test %d: ", k)
		hash := core.NewHash(fixture.hash)
		content, _ := base64.StdEncoding.DecodeString(fixture.content)
		buffer := new(bytes.Buffer)

		// Write the data out to the buffer
		testWriter(c, buffer, hash, fixture.t, content, comment)

		// Read the data back in from the buffer to be sure it matches
		testReader(c, buffer, hash, fixture.t, content, comment)
	}
}

func testWriter(c *C, dest io.Writer, hash core.Hash, typ core.ObjectType, content []byte, comment string) {
	length := int64(len(content))
	w, err := NewWriter(dest, typ, length)
	c.Assert(err, IsNil)
	c.Assert(w.Type(), Equals, typ)
	c.Assert(w.Size(), Equals, length)
	written, err := io.Copy(w, bytes.NewReader(content))
	c.Assert(err, IsNil)
	c.Assert(written, Equals, length)
	c.Assert(w.Size(), Equals, int64(len(content)))
	c.Assert(w.Hash(), Equals, hash) // Test Hash() before close
	c.Assert(w.Close(), IsNil)
	c.Assert(w.Hash(), Equals, hash) // Test Hash() after close
	_, err = w.Write([]byte{1})
	c.Assert(err, Equals, ErrClosed)
}

func (s *SuiteWriter) TestWriteOverflow(c *C) {
	w, err := NewWriter(new(bytes.Buffer), core.BlobObject, 8)
	c.Assert(err, IsNil)
	_, err = w.Write([]byte("1234"))
	c.Assert(err, IsNil)
	_, err = w.Write([]byte("56789"))
	c.Assert(err, Equals, ErrOverflow)
}

func (s *SuiteWriter) TestNewWriterInvalidType(c *C) {
	var t core.ObjectType
	_, err := NewWriter(new(bytes.Buffer), t, 8)
	c.Assert(err, Equals, core.ErrInvalidType)
}

func (s *SuiteWriter) TestNewWriterInvalidSize(c *C) {
	_, err := NewWriter(new(bytes.Buffer), core.BlobObject, -1)
	c.Assert(err, Equals, ErrNegativeSize)
	_, err = NewWriter(new(bytes.Buffer), core.BlobObject, -1651860)
	c.Assert(err, Equals, ErrNegativeSize)
}
