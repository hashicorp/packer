package git

import (
	"io"
	"path"
)

const (
	startingStackSize = 8
)

// TreeWalker provides a means of walking through all of the entries in a Tree.
type TreeWalker struct {
	stack []TreeEntryIter
	base  string

	r *Repository
}

// NewTreeWalker returns a new TreeWalker for the given repository and tree.
//
// It is the caller's responsibility to call Close() when finished with the
// tree walker.
func NewTreeWalker(r *Repository, t *Tree) *TreeWalker {
	w := TreeWalker{
		stack: make([]TreeEntryIter, 0, startingStackSize),
		base:  "",
		r:     r,
	}
	w.stack = append(w.stack, *NewTreeEntryIter(t))
	return &w
}

// Next returns the next object from the tree. Objects are returned in order
// and subtrees are included. After the last object has been returned further
// calls to Next() will return io.EOF.
//
// In the current implementation any objects which cannot be found in the
// underlying repository will be skipped automatically. It is possible that this
// may change in future versions.
func (w *TreeWalker) Next() (name string, entry TreeEntry, obj Object, err error) {
	for {
		current := len(w.stack) - 1
		if current < 0 {
			// Nothing left on the stack so we're finished
			err = io.EOF
			return
		}
		if current > maxTreeDepth {
			// We're probably following bad data or some self-referencing tree
			err = ErrMaxTreeDepth
			return
		}

		entry, err = w.stack[current].Next()
		if err == io.EOF {
			// Finished with the current tree, move back up to the parent
			w.stack = w.stack[:current]
			w.base, _ = path.Split(w.base)
			w.base = path.Clean(w.base) // Remove trailing slash
			continue
		}
		if err != nil {
			return
		}

		obj, err = w.r.Object(entry.Hash)
		if err == ObjectNotFoundErr {
			// FIXME: Avoid doing this here in case the caller actually cares about
			//        missing objects.
			err = nil
			continue // ignore entries without hash (= submodule dirs)
		}

		name = path.Join(w.base, entry.Name)

		if err != nil {
			return
		}

		break
	}

	if tree, ok := obj.(*Tree); ok {
		w.stack = append(w.stack, *NewTreeEntryIter(tree))
		w.base = path.Join(w.base, entry.Name)
	}

	return
}

// Tree returns the tree that the tree walker most recently operated on.
func (w *TreeWalker) Tree() *Tree {
	current := len(w.stack) - 1
	if current < 0 {
		return nil
	}
	return w.stack[current].t
}

// Close releases any resources used by the TreeWalker.
func (w *TreeWalker) Close() {
	w.stack = nil
}
