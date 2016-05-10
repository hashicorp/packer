package fs

// A BlockDevice is the raw device that is meant to store a filesystem.
type BlockDevice interface {
	// Closes this block device. No more methods may be called on a
	// closed device.
	Close() error

	// Len returns the number of bytes in this block device.
	Len() int64

	// SectorSize returns the size of a single sector on this device.
	SectorSize() int

	// ReadAt reads data from the block device from the given
	// offset. See io.ReaderAt for more information on this function.
	ReadAt(p []byte, off int64) (n int, err error)

	// WriteAt writes data to the block device at the given offset.
	// See io.WriterAt for more information on this function.
	WriteAt(p []byte, off int64) (n int, err error)
}
