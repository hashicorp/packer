package testutils

import "os"

// CleanupFiles removes all the provided filenames.
func CleanupFiles(moreFiles ...string) {
	for _, file := range moreFiles {
		os.RemoveAll(file)
	}
}

// FileExists returns true if the filename is found.
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
