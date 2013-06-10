package packer

import (
	"crypto/sha256"
	"path/filepath"
	"sync"
)

// Cache implements a caching interface where files can be stored for
// re-use between multiple runs.
type Cache interface {
	// Lock takes a key and returns the path where the file can be written to.
	// Packer guarantees that no other process will write to this file while
	// the lock is held.
	//
	// The cache will block and wait for the lock.
	Lock(string) (string, error)

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

func (f *FileCache) Lock(key string) (string, error) {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.Lock()

	return filepath.Join(f.CacheDir, hashKey), nil
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

	return filepath.Join(f.CacheDir, hashKey), true
}

func (f *FileCache) RUnlock(key string) {
	hashKey := f.hashKey(key)
	rw := f.rwLock(hashKey)
	rw.RUnlock()
}

func (f *FileCache) hashKey(key string) string {
	sha := sha256.New()
	sha.Write([]byte(key))
	return string(sha.Sum(nil))
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
