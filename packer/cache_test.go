package packer

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileCache_Implements(t *testing.T) {
	var raw interface{}
	raw = &FileCache{}
	if _, ok := raw.(Cache); !ok {
		t.Fatal("FileCache must be a Cache")
	}
}

func TestFileCache(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("error creating temporary dir: %s", err)
	}
	defer os.RemoveAll(cacheDir)

	cache := &FileCache{CacheDir: cacheDir}
	path := cache.Lock("foo")
	err = ioutil.WriteFile(path, []byte("data"), 0666)
	if err != nil {
		t.Fatalf("error writing: %s", err)
	}

	cache.Unlock("foo")

	path, ok := cache.RLock("foo")
	if !ok {
		t.Fatal("cache says key doesn't exist")
	}
	defer cache.RUnlock("foo")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading file: %s", err)
	}

	if string(data) != "data" {
		t.Fatalf("unknown data: %s", data)
	}
}
