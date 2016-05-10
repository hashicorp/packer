package fs

import "io"

// File is a single file within a filesystem.
type File interface {
	io.Reader
	io.Writer
}
