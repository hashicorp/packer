package packer

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type TestCache struct{}

func (TestCache) Lock(string) string {
	return ""
}

func (TestCache) Unlock(string) {}

func (TestCache) RLock(string) (string, bool) {
	return "", false
}

func (TestCache) RUnlock(string) {}

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

	// Test path with no extension (GH-716)
	path := cache.Lock("/foo.bar/baz")
	defer cache.Unlock("/foo.bar/baz")
	if strings.Contains(path, ".bar") {
		t.Fatalf("bad: %s", path)
	}

	// Test paths with a ?
	path = cache.Lock("foo.ext?foo=bar.foo")
	defer cache.Unlock("foo.ext?foo=bar.foo")
	if !strings.HasSuffix(path, ".ext") {
		t.Fatalf("bad extension with question mark: %s", path)
	}

	// Test normal paths
	path = cache.Lock("foo.iso")
	if !strings.HasSuffix(path, ".iso") {
		t.Fatalf("path doesn't end with suffix '%s': '%s'", ".iso", path)
	}

	err = ioutil.WriteFile(path, []byte("data"), 0666)
	if err != nil {
		t.Fatalf("error writing: %s", err)
	}

	cache.Unlock("foo.iso")

	path, ok := cache.RLock("foo.iso")
	if !ok {
		t.Fatal("cache says key doesn't exist")
	}
	defer cache.RUnlock("foo.iso")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading file: %s", err)
	}

	if string(data) != "data" {
		t.Fatalf("unknown data: %s", data)
	}
}
