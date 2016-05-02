package memory

import (
	"fmt"

	"gopkg.in/src-d/go-git.v3/core"
)

var ErrUnsupportedObjectType = fmt.Errorf("unsupported object type")

// ObjectStorage is the implementation of core.ObjectStorage for memory.Object
type ObjectStorage struct {
	Objects map[core.Hash]core.Object
	Commits map[core.Hash]core.Object
	Trees   map[core.Hash]core.Object
	Blobs   map[core.Hash]core.Object
	Tags    map[core.Hash]core.Object
}

// NewObjectStorage returns a new empty ObjectStorage
func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		Objects: make(map[core.Hash]core.Object, 0),
		Commits: make(map[core.Hash]core.Object, 0),
		Trees:   make(map[core.Hash]core.Object, 0),
		Blobs:   make(map[core.Hash]core.Object, 0),
		Tags:    make(map[core.Hash]core.Object, 0),
	}
}

// New returns a new empty memory.Object
func (o *ObjectStorage) New() (core.Object, error) {
	return &Object{}, nil
}

// Set stores an object, the object should be properly filled before set it.
func (o *ObjectStorage) Set(obj core.Object) (core.Hash, error) {
	h := obj.Hash()
	o.Objects[h] = obj

	switch obj.Type() {
	case core.CommitObject:
		o.Commits[h] = o.Objects[h]
	case core.TreeObject:
		o.Trees[h] = o.Objects[h]
	case core.BlobObject:
		o.Blobs[h] = o.Objects[h]
	case core.TagObject:
		o.Tags[h] = o.Objects[h]
	default:
		return h, ErrUnsupportedObjectType
	}

	return h, nil
}

// Get returns a object with the given hash
func (o *ObjectStorage) Get(h core.Hash) (core.Object, error) {
	obj, ok := o.Objects[h]
	if !ok {
		return nil, core.ObjectNotFoundErr
	}

	return obj, nil
}

// Iter returns a core.ObjectIter for the given core.ObjectTybe
func (o *ObjectStorage) Iter(t core.ObjectType) core.ObjectIter {
	var series []core.Object
	switch t {
	case core.CommitObject:
		series = flattenObjectMap(o.Commits)
	case core.TreeObject:
		series = flattenObjectMap(o.Trees)
	case core.BlobObject:
		series = flattenObjectMap(o.Blobs)
	case core.TagObject:
		series = flattenObjectMap(o.Tags)
	}
	return core.NewObjectSliceIter(series)
}

func flattenObjectMap(m map[core.Hash]core.Object) []core.Object {
	objects := make([]core.Object, 0, len(m))
	for _, obj := range m {
		objects = append(objects, obj)
	}
	return objects
}
