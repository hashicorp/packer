package common

import (
        "os"
        "time"
)

type MockFileInfo struct { fileName string }
func (fi *MockFileInfo) Name() string {
        return fi.fileName
}
func (fi *MockFileInfo) IsDir() bool {
        return false
}
func (fi *MockFileInfo) ModTime() time.Time {
        return time.Now()
}
func (fi *MockFileInfo) Mode() os.FileMode {
        return 0777
}
func (fi *MockFileInfo) Size() int64 {
        return 0
}
func (fi *MockFileInfo) Sys() interface{} {
        return nil
}
