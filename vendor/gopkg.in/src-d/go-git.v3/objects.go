package git

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/src-d/go-git.v3/core"
)

var ErrUnsupportedObject = errors.New("unsupported object type")

// Object is a generic representation of any git object. It is implemented by
// Commit, Tree, Blob and Tag, and includes the functions that are common to
// them.
//
// Object is returned when an object could of any type. It is frequently used
// with a type cast to acquire the specific type of object:
//
//   func process(obj Object) {
//   	switch o := obj.(type) {
//   	case *Commit:
//   		// o is a Commit
//   	case *Tree:
//   		// o is a Tree
//   	case *Blob:
//   		// o is a Blob
//   	case *Tag:
//   		// o is a Tag
//   	}
//   }
//
// This interface is intentionally different from core.Object, which is a lower
// level interface used by storage implementations to read and write objects.
type Object interface {
	ID() core.Hash
	Type() core.ObjectType
	Decode(core.Object) error
}

// Blob is used to store file data - it is generally a file.
type Blob struct {
	Hash core.Hash
	Size int64

	obj core.Object
}

// ID returns the object ID of the blob. The returned value will always match
// the current value of Blob.Hash.
//
// ID is present to fufill the Object interface.
func (b *Blob) ID() core.Hash {
	return b.Hash
}

// Type returns the type of object. It always returns core.BlobObject.
//
// Type is present to fufill the Object interface.
func (b *Blob) Type() core.ObjectType {
	return core.BlobObject
}

// Decode transforms a core.Object into a Blob struct.
func (b *Blob) Decode(o core.Object) error {
	if o.Type() != core.BlobObject {
		return ErrUnsupportedObject
	}

	b.Hash = o.Hash()
	b.Size = o.Size()
	b.obj = o

	return nil
}

// Reader returns a reader allow the access to the content of the blob
func (b *Blob) Reader() (core.ObjectReader, error) {
	return b.obj.Reader()
}

// Signature represents an action signed by a person
type Signature struct {
	Name  string
	Email string
	When  time.Time
}

// Decode decodes a byte slice into a signature
func (s *Signature) Decode(b []byte) {
	open := bytes.IndexByte(b, '<')
	close := bytes.IndexByte(b, '>')
	if open == -1 || close == -1 {
		return
	}

	s.Name = string(bytes.Trim(b[:open], " "))
	s.Email = string(b[open+1 : close])

	hasTime := close+2 < len(b)
	if hasTime {
		s.decodeTimeAndTimeZone(b[close+2:])
	}
}

var timeZoneLength = 5

func (s *Signature) decodeTimeAndTimeZone(b []byte) {
	space := bytes.IndexByte(b, ' ')
	if space == -1 {
		space = len(b)
	}

	ts, err := strconv.ParseInt(string(b[:space]), 10, 64)
	if err != nil {
		return
	}

	s.When = time.Unix(ts, 0).In(time.UTC)
	var tzStart = space + 1
	if tzStart >= len(b) || tzStart+timeZoneLength > len(b) {
		return
	}

	tl, err := time.Parse("-0700", string(b[tzStart:tzStart+timeZoneLength]))
	if err != nil {
		return
	}

	s.When = s.When.In(tl.Location())
}

func (s *Signature) String() string {
	return fmt.Sprintf("%s <%s>", s.Name, s.Email)
}
