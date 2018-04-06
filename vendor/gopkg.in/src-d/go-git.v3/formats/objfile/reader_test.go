package objfile

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/core"

	"github.com/klauspost/compress/zlib"
)

type SuiteReader struct{}

var _ = Suite(&SuiteReader{})

func (s *SuiteReader) TestReadObjfile(c *C) {
	for k, fixture := range objfileFixtures {
		comment := fmt.Sprintf("test %d: ", k)
		hash := core.NewHash(fixture.hash)
		content, _ := base64.StdEncoding.DecodeString(fixture.content)
		data, _ := base64.StdEncoding.DecodeString(fixture.data)

		testReader(c, bytes.NewReader(data), hash, fixture.t, content, comment)
	}
}

func testReader(c *C, source io.Reader, hash core.Hash, typ core.ObjectType, content []byte, comment string) {
	r, err := NewReader(source)
	c.Assert(err, IsNil)
	c.Assert(r.Type(), Equals, typ)
	rc, err := ioutil.ReadAll(r)
	c.Assert(err, IsNil)
	c.Assert(rc, DeepEquals, content, Commentf("%scontent=%s, expected=%s", base64.StdEncoding.EncodeToString(rc), base64.StdEncoding.EncodeToString(content)))
	c.Assert(r.Size(), Equals, int64(len(content)))
	c.Assert(r.Hash(), Equals, hash) // Test Hash() before close
	c.Assert(r.Close(), IsNil)
	c.Assert(r.Hash(), Equals, hash) // Test Hash() after close
	_, err = r.Read(make([]byte, 0, 1))
	c.Assert(err, Equals, ErrClosed)
}

func (s *SuiteReader) TestReadEmptyObjfile(c *C) {
	source := bytes.NewReader([]byte{})
	_, err := NewReader(source)
	c.Assert(err, Equals, ErrZLib)
}

func (s *SuiteReader) TestReadEmptyContent(c *C) {
	b := new(bytes.Buffer)
	w := zlib.NewWriter(b)
	c.Assert(w.Close(), IsNil)
	_, err := NewReader(b)
	c.Assert(err, Equals, ErrHeader)
}

func (s *SuiteReader) TestReadGarbage(c *C) {
	source := bytes.NewReader([]byte("!@#$RO!@NROSADfinq@o#irn@oirfn"))
	_, err := NewReader(source)
	c.Assert(err, Equals, ErrZLib)
}

func (s *SuiteReader) TestReadCorruptZLib(c *C) {
	data, _ := base64.StdEncoding.DecodeString("eAFLysaalPUjBgAAAJsAHw")
	source := bytes.NewReader(data)
	_, err := NewReader(source)
	c.Assert(err, NotNil)
}
