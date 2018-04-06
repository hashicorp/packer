// Package core implement the core interfaces and structs used by go-git
package core

import (
	"errors"
	"io"
)

var (
	ObjectNotFoundErr = errors.New("object not found")
	// ErrInvalidType is returned when an invalid object type is provided.
	ErrInvalidType = errors.New("invalid object type")
)

// TODO: Consider adding a Hash function to the ObjectReader and ObjectWriter
//       interfaces that returns the hash calculated for the reader or writer.

// ObjectReader is a generic representation of an object reader.
//
// ObjectReader implements io.ReadCloser. Close should be called when finished
// with it.
type ObjectReader io.ReadCloser

// ObjectWriter is a generic representation of an object writer.
//
// ObjectWriter implements io.WriterCloser. Close should be called when finished
// with it.
type ObjectWriter io.WriteCloser

// Object is a generic representation of any git object
type Object interface {
	Hash() Hash
	Type() ObjectType
	SetType(ObjectType)
	Size() int64
	SetSize(int64)
	Reader() (ObjectReader, error)
	Writer() (ObjectWriter, error)
}

// ObjectStorage generic storage of objects
type ObjectStorage interface {
	New() (Object, error)
	Set(Object) (Hash, error)
	Get(Hash) (Object, error)
	Iter(ObjectType) ObjectIter
}

// ObjectType internal object type's
type ObjectType int8

const (
	CommitObject   ObjectType = 1
	TreeObject     ObjectType = 2
	BlobObject     ObjectType = 3
	TagObject      ObjectType = 4
	OFSDeltaObject ObjectType = 6
	REFDeltaObject ObjectType = 7
)

func (t ObjectType) String() string {
	switch t {
	case CommitObject:
		return "commit"
	case TreeObject:
		return "tree"
	case BlobObject:
		return "blob"
	case TagObject:
		return "tag"
	case OFSDeltaObject:
		return "ofs-delta"
	case REFDeltaObject:
		return "ref-delta"
	default:
		return "unknown"
	}
}

func (t ObjectType) Bytes() []byte {
	return []byte(t.String())
}

// Valid returns true if t is a valid ObjectType.
func (t ObjectType) Valid() bool {
	return t >= CommitObject && t <= REFDeltaObject
}

// ParseObjectType parses a string representation of ObjectType. It returns an
// error on parse failure.
func ParseObjectType(value string) (typ ObjectType, err error) {
	switch value {
	case "commit":
		typ = CommitObject
	case "tree":
		typ = TreeObject
	case "blob":
		typ = BlobObject
	case "tag":
		typ = TagObject
	case "ofs-delta":
		typ = OFSDeltaObject
	case "ref-delta":
		typ = REFDeltaObject
	default:
		err = ErrInvalidType
	}
	return
}

// ObjectIter is a generic closable interface for iterating over objects.
type ObjectIter interface {
	Next() (Object, error)
	Close()
}

// ObjectLookupIter implements ObjectIter. It iterates over a series of object
// hashes and yields their associated objects by retrieving each one from
// object storage. The retrievals are lazy and only occur when the iterator
// moves forward with a call to Next().
//
// The ObjectLookupIter must be closed with a call to Close() when it is no
// longer needed.
type ObjectLookupIter struct {
	storage ObjectStorage
	series  []Hash
	pos     int
}

// NewObjectLookupIter returns an object iterator given an object storage and
// a slice of object hashes.
func NewObjectLookupIter(storage ObjectStorage, series []Hash) *ObjectLookupIter {
	return &ObjectLookupIter{
		storage: storage,
		series:  series,
	}
}

// Next returns the next object from the iterator. If the iterator has reached
// the end it will return io.EOF as an error. If the object can't be found in
// the object storage, it will return ObjectNotFoundErr as an error. If the
// object is retreieved successfully error will be nil.
func (iter *ObjectLookupIter) Next() (Object, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}
	hash := iter.series[iter.pos]
	obj, err := iter.storage.Get(hash)
	if err == nil {
		iter.pos++
	}
	return obj, err
}

// Close releases any resources used by the iterator.
func (iter *ObjectLookupIter) Close() {
	iter.pos = len(iter.series)
}

// ObjectSliceIter implements ObjectIter. It iterates over a series of objects
// stored in a slice and yields each one in turn when Next() is called.
//
// The ObjectSliceIter must be closed with a call to Close() when it is no
// longer needed.
type ObjectSliceIter struct {
	series []Object
	pos    int
}

// NewObjectSliceIter returns an object iterator for the given slice of objects.
func NewObjectSliceIter(series []Object) *ObjectSliceIter {
	return &ObjectSliceIter{
		series: series,
	}
}

// Next returns the next object from the iterator. If the iterator has reached
// the end it will return io.EOF as an error. If the object is retreieved
// successfully error will be nil.
func (iter *ObjectSliceIter) Next() (Object, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}
	obj := iter.series[iter.pos]
	iter.pos++
	return obj, nil
}

// Close releases any resources used by the iterator.
func (iter *ObjectSliceIter) Close() {
	iter.pos = len(iter.series)
}
