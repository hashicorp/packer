package git

import (
	"bytes"
	"os"
	"strings"
)

// File represents git file objects.
type File struct {
	Name string
	Mode os.FileMode
	Blob
}

func newFile(name string, m os.FileMode, b *Blob) *File {
	return &File{Name: name, Mode: m, Blob: *b}
}

// Contents returns the contents of a file as a string.
func (f *File) Contents() (content string, err error) {
	reader, err := f.Reader()
	if err != nil {
		return "", err
	}
	defer checkClose(reader, &err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	return buf.String(), nil
}

// Lines returns a slice of lines from the contents of a file, stripping
// all end of line characters. If the last line is empty (does not end
// in an end of line), it is also stripped.
func (f *File) Lines() ([]string, error) {
	content, err := f.Contents()
	if err != nil {
		return nil, err
	}

	splits := strings.Split(content, "\n")
	// remove the last line if it is empty
	if splits[len(splits)-1] == "" {
		return splits[:len(splits)-1], nil
	}

	return splits, nil
}

type FileIter struct {
	w TreeWalker
}

func NewFileIter(r *Repository, t *Tree) *FileIter {
	return &FileIter{w: *NewTreeWalker(r, t)}
}

func (iter *FileIter) Next() (*File, error) {
	for {
		name, entry, obj, err := iter.w.Next()
		if err != nil {
			return nil, err
		}

		if blob, ok := obj.(*Blob); ok {
			return newFile(name, entry.Mode, blob), nil
		}
	}
}

func (iter *FileIter) Close() {
	iter.w.Close()
}
