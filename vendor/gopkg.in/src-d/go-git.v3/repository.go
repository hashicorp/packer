package git

import (
	"errors"
	"fmt"

	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/formats/packfile"
	"gopkg.in/src-d/go-git.v3/storage/memory"
)

var (
	ObjectNotFoundErr = errors.New("object not found")
)

const (
	DefaultRemoteName = "origin"
)

type Repository struct {
	Remotes map[string]*Remote
	Storage core.ObjectStorage
	URL     string
}

// NewRepository creates a new repository setting remote as default remote
func NewRepository(url string, auth common.AuthMethod) (*Repository, error) {
	var remote *Remote
	var err error

	if auth == nil {
		remote, err = NewRemote(url)
	} else {
		remote, err = NewAuthenticatedRemote(url, auth)
	}

	if err != nil {
		return nil, err
	}

	r := NewPlainRepository()
	r.Remotes[DefaultRemoteName] = remote
	r.URL = url

	return r, nil
}

// NewPlainRepository creates a new repository without remotes
func NewPlainRepository() *Repository {
	return &Repository{
		Remotes: map[string]*Remote{},
		Storage: memory.NewObjectStorage(),
	}
}

// Pull connect and fetch the given branch from the given remote, the branch
// should be provided with the full path not only the abbreviation, eg.:
// "refs/heads/master"
func (r *Repository) Pull(remoteName, branch string) (err error) {
	remote, ok := r.Remotes[remoteName]
	if !ok {
		return fmt.Errorf("unable to find remote %q", remoteName)
	}

	if err := remote.Connect(); err != nil {
		return err
	}

	if branch == "" {
		branch = remote.DefaultBranch()
	}

	ref, err := remote.Ref(branch)
	if err != nil {
		return err
	}

	req := &common.GitUploadPackRequest{}
	req.Want(ref)

	// TODO: Provide "haves" for what's already in the repository's storage

	reader, err := remote.Fetch(req)
	if err != nil {
		return err
	}
	defer checkClose(reader, &err)

	pr := packfile.NewReader(reader)
	if _, err = pr.Read(r.Storage); err != nil {
		return err
	}

	return nil
}

// PullDefault like Pull but retrieve the default branch from the default remote
func (r *Repository) PullDefault() (err error) {
	return r.Pull(DefaultRemoteName, "")
}

// Commit return the commit with the given hash
func (r *Repository) Commit(h core.Hash) (*Commit, error) {
	obj, err := r.Storage.Get(h)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ObjectNotFoundErr
		}
		return nil, err
	}

	commit := &Commit{r: r}
	return commit, commit.Decode(obj)
}

// Commits decode the objects into commits
func (r *Repository) Commits() *CommitIter {
	return NewCommitIter(r, r.Storage.Iter(core.CommitObject))
}

// Tree return the tree with the given hash
func (r *Repository) Tree(h core.Hash) (*Tree, error) {
	obj, err := r.Storage.Get(h)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ObjectNotFoundErr
		}
		return nil, err
	}

	tree := &Tree{r: r}
	return tree, tree.Decode(obj)
}

// Blob returns the blob with the given hash
func (r *Repository) Blob(h core.Hash) (*Blob, error) {
	obj, err := r.Storage.Get(h)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ObjectNotFoundErr
		}
		return nil, err
	}

	blob := &Blob{}
	return blob, blob.Decode(obj)
}

// Tag returns a tag with the given hash.
func (r *Repository) Tag(h core.Hash) (*Tag, error) {
	obj, err := r.Storage.Get(h)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ObjectNotFoundErr
		}
		return nil, err
	}

	tag := &Tag{r: r}
	return tag, tag.Decode(obj)
}

// Tags returns a TagIter that can step through all of the annotated tags
// in the repository.
func (r *Repository) Tags() *TagIter {
	return NewTagIter(r, r.Storage.Iter(core.TagObject))
}

// Object returns an object with the given hash.
func (r *Repository) Object(h core.Hash) (Object, error) {
	obj, err := r.Storage.Get(h)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return nil, ObjectNotFoundErr
		}
		return nil, err
	}

	switch obj.Type() {
	case core.CommitObject:
		commit := &Commit{r: r}
		return commit, commit.Decode(obj)
	case core.TreeObject:
		tree := &Tree{r: r}
		return tree, tree.Decode(obj)
	case core.BlobObject:
		blob := &Blob{}
		return blob, blob.Decode(obj)
	case core.TagObject:
		tag := &Tag{r: r}
		return tag, tag.Decode(obj)
	default:
		return nil, core.ErrInvalidType
	}
}
