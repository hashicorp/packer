package memory

import (
	"bytes"
	"io/ioutil"

	"gopkg.in/src-d/go-git.v3/core"
)

// Object on memory core.Object implementation
type Object struct {
	t       core.ObjectType
	h       core.Hash
	content []byte
	size    int64
}

// Hash return the object Hash, the hash is calculated on-the-fly the first
// time is called, the subsequent calls the same Hash is returned even in the
// type or the content has changed. The Hash is only generated if the size of
// the content is exactly the Object.Size
func (o *Object) Hash() core.Hash {
	if o.h == core.ZeroHash && int64(len(o.content)) == o.size {
		o.h = core.ComputeHash(o.t, o.content)
	}

	return o.h
}

// Type return the core.ObjectType
func (o *Object) Type() core.ObjectType { return o.t }

// SetType sets the core.ObjectType
func (o *Object) SetType(t core.ObjectType) { o.t = t }

// Size return the size of the object
func (o *Object) Size() int64 { return o.size }

// SetSize set the object size, the given size should be written afterwards
func (o *Object) SetSize(s int64) { o.size = s }

// Reader returns a core.ObjectReader used to read the object's content.
func (o *Object) Reader() (core.ObjectReader, error) {
	return ioutil.NopCloser(bytes.NewBuffer(o.content)), nil
}

// Writer returns a core.ObjectWriter used to write the object's content.
func (o *Object) Writer() (core.ObjectWriter, error) {
	return o, nil
}

func (o *Object) Write(p []byte) (n int, err error) {
	o.content = append(o.content, p...)
	return len(p), nil
}

// Close releases any resources consumed by the object when it is acting as a
// core.ObjectWriter.
func (o *Object) Close() error { return nil }
