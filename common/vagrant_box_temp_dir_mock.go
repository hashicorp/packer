package common

import (
        "errors"
        "os"
)

type MockVagrantBoxTempDir struct {
        path         string
        errorMessage string
        fileInfo     os.FileInfo
}
func (vbtd *MockVagrantBoxTempDir) Create() (error) {
        vbtd.path = "don't care"
        if vbtd.errorMessage != "" {
                return errors.New(vbtd.errorMessage)
        }
        return nil
}
func (vbtd *MockVagrantBoxTempDir) Path() string {
        return vbtd.path
}
func (vbtd *MockVagrantBoxTempDir) ReadDir(targetPath string) ([]os.FileInfo, error) {
        array := []os.FileInfo{}
        if vbtd.fileInfo != nil {
          array = append(array, vbtd.fileInfo)
        }
        return array, nil
}
func (vbtd *MockVagrantBoxTempDir) FindFileWithSuffix(suffix string) os.FileInfo {
        if vbtd.fileInfo != nil {
          return vbtd.fileInfo
        }
        return nil
}
