package packer

import (
	"os"
	"path/filepath"
)

var DefaultCacheDir = "packer_cache"

// CachePath returns an absolute path to a cache file or directory
//
// When the directory is not absolute, CachePath will try to get
// current working directory to be able to return a full path.
// CachePath tries to create the resulting path if it doesn't exist.
//
// CachePath can error in case it cannot find the cwd.
//
// ex:
//   PACKER_CACHE_DIR=""            CacheDir() => "./packer_cache/
//   PACKER_CACHE_DIR=""            CacheDir("foo") => "./packer_cache/foo
//   PACKER_CACHE_DIR="bar"         CacheDir("foo") => "./bar/foo
//   PACKER_CACHE_DIR="/home/there" CacheDir("foo", "bar") => "/home/there/foo/bar
func CachePath(paths ...string) (path string, err error) {
	defer func() {
		// create the dir based on return path if it doesn't exist
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
	}()
	cacheDir := DefaultCacheDir
	if cd := os.Getenv("PACKER_CACHE_DIR"); cd != "" {
		cacheDir = cd
	}

	paths = append([]string{cacheDir}, paths...)
	return filepath.Abs(filepath.Join(paths...))
}
