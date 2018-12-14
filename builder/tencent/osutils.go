package tencent

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

// generates a temporary file name in the temporary directory
func TempFileName() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), hex.EncodeToString(randBytes))
}

// See https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go/12518877#12518877
func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// DirectoryExists tests that the given path exists
func DirectoryExists(path string) bool {
	if path == "" {
		return true
	}
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || !stat.IsDir() {
		return false
	}

	return true
}

// FileExists tests that the given filename exists
func FileExists(path string) bool {
	return pathExists(path)
}
