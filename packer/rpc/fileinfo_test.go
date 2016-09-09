package rpc

import (
	"os"
	"testing"
	"time"
)

type dummyFileInfo struct{}

func (fi dummyFileInfo) Name() string      { return "dummy" }
func (fi dummyFileInfo) Size() int64       { return 64 }
func (fi dummyFileInfo) Mode() os.FileMode { return 0644 }
func (fi dummyFileInfo) ModTime() time.Time {
	return time.Time{}.Add(1 * time.Minute)
}
func (fi dummyFileInfo) IsDir() bool      { return false }
func (fi dummyFileInfo) Sys() interface{} { return nil }
func TestNewFileInfoNilPointer(t *testing.T) {
	fi := NewFileInfo(os.FileInfo(nil))
	if fi != nil {
		t.Fatalf("should be nil")
	}
}

func TestNewFileInfoValues(t *testing.T) {
	in := dummyFileInfo{}
	fi := NewFileInfo(in)

	if fi.Size() != in.Size() {
		t.Errorf("fi.Size() = %d; want %d", fi.Size(), in.Size())
	}

	if fi.Name() != in.Name() {
		t.Errorf("fi.Name() = %s; want %s", fi.Name(), in.Name())
	}

	if fi.Mode() != in.Mode() {
		t.Errorf("fi.Mode() = %#o; want %#o", fi.Mode(), in.Mode())
	}

	if fi.ModTime() != in.ModTime() {
		t.Errorf("fi.ModTime() = %s; want %s", fi.ModTime(), in.ModTime())
	}

	if fi.IsDir() != in.IsDir() {
		t.Errorf("fi.IsDir() = %t; want %t", fi.IsDir(), in.IsDir())
	}
}
