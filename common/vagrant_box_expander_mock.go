package common

import (
        "errors"
)

type MockVagrantBoxExpander struct {
        errorMessage string
}
func (vbe *MockVagrantBoxExpander) Expand(sourcePath string, targetPath string) (error) {
        if vbe.errorMessage != "" {
                return errors.New(vbe.errorMessage)
        }
        return nil
}
