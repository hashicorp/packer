package rpc

import (
	"os"
	"time"
)

func NewFileInfo(name string, size int64, mode os.FileMode, mtime time.Time) *fileInfo {
	return &fileInfo{N: name, S: size, M: mode, T: mtime}
}

type fileInfo struct {
	N string
	S int64
	M os.FileMode
	T time.Time
}

func (fi fileInfo) Name() string      { return fi.N }
func (fi fileInfo) Size() int64       { return fi.S }
func (fi fileInfo) Mode() os.FileMode { return fi.M }
func (fi fileInfo) ModTime() time.Time {
	if fi.T.IsZero() {
		return time.Now()
	}
	return fi.T
}
func (fi fileInfo) IsDir() bool      { return fi.M.IsDir() }
func (fi fileInfo) Sys() interface{} { return nil }
