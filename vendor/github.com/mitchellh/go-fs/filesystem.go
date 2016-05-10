package fs

// A FileSystem provides access to a tree hierarchy of directories
// and files.
type FileSystem interface {
	// RootDir returns the single root directory.
	RootDir() (Directory, error)
}
