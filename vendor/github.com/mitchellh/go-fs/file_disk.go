package fs

import (
	"errors"
	"os"
)

// A FileDisk is an implementation of a BlockDevice that uses a
// *os.File as its backing store.
type FileDisk struct {
	f    *os.File
	size int64
}

// NewFileDisk creates a new FileDisk from the given *os.File. The
// file must already be created and set the to the proper size.
func NewFileDisk(f *os.File) (*FileDisk, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, errors.New("file is a directory")
	}

	return &FileDisk{
		f:    f,
		size: fi.Size(),
	}, nil
}

func (f *FileDisk) Close() error {
	return f.f.Close()
}

func (f *FileDisk) Len() int64 {
	return f.size
}

func (f *FileDisk) ReadAt(p []byte, off int64) (int, error) {
	return f.f.ReadAt(p, off)
}

func (f *FileDisk) SectorSize() int {
	// Hardcoded for now, one day we may want to make this customizable
	return 512
}

func (f *FileDisk) WriteAt(p []byte, off int64) (int, error) {
	return f.f.WriteAt(p, off)
}
