package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"

	"gopkg.in/src-d/go-git.v3/core"
)

type Hash core.Hash

// Commit points to a single tree, marking it as what the project looked like
// at a certain point in time. It contains meta-information about that point
// in time, such as a timestamp, the author of the changes since the last
// commit, a pointer to the previous commit(s), etc.
// http://schacon.github.io/gitbook/1_the_git_object_model.html
type Commit struct {
	Hash      core.Hash
	Author    Signature
	Committer Signature
	Message   string

	tree    core.Hash
	parents []core.Hash
	r       *Repository
}

func (c *Commit) Tree() *Tree {
	tree, _ := c.r.Tree(c.tree) // FIXME: Return error as well?
	return tree
}

func (c *Commit) Parents() *CommitIter {
	return NewCommitIter(c.r, core.NewObjectLookupIter(c.r.Storage, c.parents))
}

// NumParents returns the number of parents in a commit.
func (c *Commit) NumParents() int {
	return len(c.parents)
}

// File returns the file with the specified "path" in the commit and a
// nil error if the file exists. If the file does not exist, it returns
// a nil file and the ErrFileNotFound error.
func (c *Commit) File(path string) (file *File, err error) {
	return c.Tree().File(path)
}

// ID returns the object ID of the commit. The returned value will always match
// the current value of Commit.Hash.
//
// ID is present to fufill the Object interface.
func (c *Commit) ID() core.Hash {
	return c.Hash
}

// Type returns the type of object. It always returns core.CommitObject.
//
// Type is present to fufill the Object interface.
func (c *Commit) Type() core.ObjectType {
	return core.CommitObject
}

// Decode transforms a core.Object into a Commit struct.
func (c *Commit) Decode(o core.Object) (err error) {
	if o.Type() != core.CommitObject {
		return ErrUnsupportedObject
	}

	c.Hash = o.Hash()

	reader, err := o.Reader()
	if err != nil {
		return err
	}
	defer checkClose(reader, &err)

	r := bufio.NewReader(reader)

	var message bool
	for {
		line, err := r.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return err
		}

		line = bytes.TrimSpace(line)
		if !message {
			if len(line) == 0 {
				message = true
				continue
			}

			split := bytes.SplitN(line, []byte{' '}, 2)
			switch string(split[0]) {
			case "tree":
				c.tree = core.NewHash(string(split[1]))
			case "parent":
				c.parents = append(c.parents, core.NewHash(string(split[1])))
			case "author":
				c.Author.Decode(split[1])
			case "committer":
				c.Committer.Decode(split[1])
			}
		} else {
			c.Message += string(line) + "\n"
		}

		if err == io.EOF {
			return nil
		}
	}
}

func (c *Commit) String() string {
	return fmt.Sprintf(
		"%s %s\nAuthor: %s\nDate:   %s\n",
		core.CommitObject, c.Hash, c.Author.String(), c.Author.When,
	)
}

// CommitIter provides an iterator for a set of commits.
type CommitIter struct {
	core.ObjectIter
	r *Repository
}

// NewCommitIter returns a CommitIter for the given repository and underlying
// object iterator.
//
// The returned CommitIter will automatically skip over non-commit objects.
func NewCommitIter(r *Repository, iter core.ObjectIter) *CommitIter {
	return &CommitIter{iter, r}
}

// Next moves the iterator to the next commit and returns a pointer to it. If it
// has reached the end of the set it will return io.EOF.
func (iter *CommitIter) Next() (*Commit, error) {
	obj, err := iter.ObjectIter.Next()
	if err != nil {
		return nil, err
	}

	commit := &Commit{r: iter.r}
	return commit, commit.Decode(obj)
}

type commitSorterer struct {
	l []*Commit
}

func (s commitSorterer) Len() int {
	return len(s.l)
}

func (s commitSorterer) Less(i, j int) bool {
	return s.l[i].Committer.When.Before(s.l[j].Committer.When)
}

func (s commitSorterer) Swap(i, j int) {
	s.l[i], s.l[j] = s.l[j], s.l[i]
}

// SortCommits sort a commit list by commit date, from older to newer.
func SortCommits(l []*Commit) {
	s := &commitSorterer{l}
	sort.Sort(s)
}
