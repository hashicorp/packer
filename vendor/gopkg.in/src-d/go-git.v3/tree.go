package git

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/src-d/go-git.v3/core"
)

const (
	maxTreeDepth = 1024
)

// New errors defined by this package.
var (
	ErrMaxTreeDepth = errors.New("maximum tree depth exceeded")
	ErrFileNotFound = errors.New("file not found")
)

// Tree is basically like a directory - it references a bunch of other trees
// and/or blobs (i.e. files and sub-directories)
type Tree struct {
	Entries []TreeEntry
	Hash    core.Hash

	r *Repository
	m map[string]*TreeEntry
}

// TreeEntry represents a file
type TreeEntry struct {
	Name string
	Mode os.FileMode
	Hash core.Hash
}

// File returns the hash of the file identified by the `path` argument.
// The path is interpreted as relative to the tree receiver.
func (t *Tree) File(path string) (*File, error) {
	e, err := t.findEntry(path)
	if err != nil {
		return nil, ErrFileNotFound
	}

	obj, err := t.r.Storage.Get(e.Hash)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ErrFileNotFound // a git submodule
		}
		return nil, err
	}

	if obj.Type() != core.BlobObject {
		return nil, ErrFileNotFound // a directory
	}

	blob := &Blob{}
	blob.Decode(obj)

	return newFile(path, e.Mode, blob), nil
}

func (t *Tree) findEntry(path string) (*TreeEntry, error) {
	pathParts := strings.Split(path, "/")

	var tree *Tree
	var err error
	for tree = t; len(pathParts) > 1; pathParts = pathParts[1:] {
		if tree, err = tree.dir(pathParts[0]); err != nil {
			return nil, err
		}
	}

	return tree.entry(pathParts[0])
}

var errDirNotFound = errors.New("directory not found")

func (t *Tree) dir(baseName string) (*Tree, error) {
	entry, err := t.entry(baseName)
	if err != nil {
		return nil, errDirNotFound
	}

	obj, err := t.r.Storage.Get(entry.Hash)
	if err != nil {
		if err == core.ObjectNotFoundErr { // git submodule
			return nil, errDirNotFound
		}
		return nil, err
	}

	if obj.Type() != core.TreeObject {
		return nil, errDirNotFound // a file
	}

	tree := &Tree{r: t.r}
	tree.Decode(obj)

	return tree, nil
}

var errEntryNotFound = errors.New("entry not found")

func (t *Tree) entry(baseName string) (*TreeEntry, error) {
	if t.m == nil {
		t.buildMap()
	}
	entry, ok := t.m[baseName]
	if !ok {
		return nil, errEntryNotFound
	}

	return entry, nil
}

func (t *Tree) Files() *FileIter {
	return NewFileIter(t.r, t)
}

// ID returns the object ID of the tree. The returned value will always match
// the current value of Tree.Hash.
//
// ID is present to fufill the Object interface.
func (t *Tree) ID() core.Hash {
	return t.Hash
}

// Type returns the type of object. It always returns core.TreeObject.
func (t *Tree) Type() core.ObjectType {
	return core.TreeObject
}

// Decode transform an core.Object into a Tree struct
func (t *Tree) Decode(o core.Object) (err error) {
	if o.Type() != core.TreeObject {
		return ErrUnsupportedObject
	}

	t.Hash = o.Hash()
	if o.Size() == 0 {
		return nil
	}

	t.Entries = nil
	t.m = nil

	reader, err := o.Reader()
	if err != nil {
		return err
	}
	defer checkClose(reader, &err)

	r := bufio.NewReader(reader)
	for {
		mode, err := r.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fm, err := strconv.ParseInt(mode[:len(mode)-1], 8, 32)
		if err != nil && err != io.EOF {
			return err
		}

		name, err := r.ReadString(0)
		if err != nil && err != io.EOF {
			return err
		}

		var hash core.Hash
		_, err = r.Read(hash[:])
		if err != nil && err != io.EOF {
			return err
		}

		baseName := name[:len(name)-1]
		t.Entries = append(t.Entries, TreeEntry{
			Hash: hash,
			Mode: os.FileMode(fm),
			Name: baseName,
		})
	}

	return nil
}

func (t *Tree) buildMap() {
	t.m = make(map[string]*TreeEntry)
	for i := 0; i < len(t.Entries); i++ {
		t.m[t.Entries[i].Name] = &t.Entries[i]
	}
}

// TreeEntryIter facilitates iterating through the TreeEntry objects in a Tree.
type TreeEntryIter struct {
	t   *Tree
	pos int
}

func NewTreeEntryIter(t *Tree) *TreeEntryIter {
	return &TreeEntryIter{t, 0}
}

func (iter *TreeEntryIter) Next() (TreeEntry, error) {
	if iter.pos >= len(iter.t.Entries) {
		return TreeEntry{}, io.EOF
	}
	iter.pos++
	return iter.t.Entries[iter.pos-1], nil
}

// TreeEntryIter facilitates iterating through the descendent subtrees of a
// Tree.
type TreeIter struct {
	w TreeWalker
}

func NewTreeIter(r *Repository, t *Tree) *TreeIter {
	return &TreeIter{
		w: *NewTreeWalker(r, t),
	}
}

func (iter *TreeIter) Next() (*Tree, error) {
	for {
		_, _, obj, err := iter.w.Next()
		if err != nil {
			return nil, err
		}

		if tree, ok := obj.(*Tree); ok {
			return tree, nil
		}
	}
}

func (iter *TreeIter) Close() {
	iter.w.Close()
}
