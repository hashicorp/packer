package packer

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Cache implements a caching interface where files can be stored for
// re-use between multiple runs.
type Cache interface {
	// Lock takes a key and returns the path where the file can be written to.
	// Packer guarantees that no other process will write to this file while
	// the lock is held.
	//
	// If the key has an extension (e.g., file.ext), the resulting path
	// will have that extension as well.
	//
	// The cache will block and wait for the lock.
	Lock(string) string

	// Unlock will unlock a certain cache key. Be very careful that this
	// is only called once per lock obtained.
	Unlock(string)

	// RLock returns the path to a key in the cache and locks it for reading.
	// The second return parameter is whether the key existed or not.
	// This will block if any locks are held for writing. No lock will be
	// held if the key doesn't exist.
	RLock(string) (string, bool)

	// RUnlock will unlock a key for reading.
	RUnlock(string)
}

// FileCache implements a Cache by caching the data directly to a cache
// directory.
type FileCache struct {
	CacheDir string
	l        sync.Mutex
	rw       map[string]*sync.RWMutex
}

func (f *FileCache) Lock(key string) string {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.Lock()

	return f.cachePath(key, hashKey)
}

func (f *FileCache) Unlock(key string) {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.Unlock()
}

func (f *FileCache) RLock(key string) (string, bool) {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.RLock()

	return f.cachePath(key, hashKey), true
}

func (f *FileCache) RUnlock(key string) {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.RUnlock()
}

func (f *FileCache) cachePath(key string, hashKey string) string {
	if endIndex := strings.Index(key, "?"); endIndex > -1 {
		key = key[:endIndex]
	}

	suffix := ""
	dotIndex := strings.LastIndex(key, ".")
	if dotIndex > -1 {
		if slashIndex := strings.LastIndex(key, "/"); slashIndex <= dotIndex {
			suffix = key[dotIndex:]
		}
	}

	// Make the cache directory. We ignore errors here, but
	// log them in case something happens.
	if err := os.MkdirAll(f.CacheDir, 0755); err != nil {
		log.Printf(
			"[ERR] Error making cacheDir: %s %s", f.CacheDir, err)
	}

	return filepath.Join(f.CacheDir, hashKey+suffix)
}

func (f *FileCache) hashKey(key string) string {
	sha := sha256.New()
	sha.Write([]byte(key))
	return hex.EncodeToString(sha.Sum(nil))
}

func (f *FileCache) rwLock(hashKey string) *sync.RWMutex {
	f.l.Lock()
	defer f.l.Unlock()

	if f.rw == nil {
		f.rw = make(map[string]*sync.RWMutex)
	}

	if result, ok := f.rw[hashKey]; ok {
		return result
	}

	var result sync.RWMutex
	f.rw[hashKey] = &result
	return &result
}
