package common

import (
	"os"
        "path/filepath"
	"testing"
        "github.com/stretchr/testify/assert"
)

func TestVagrantBox_ReturnPathIfPathIsNotABox(t *testing.T) {

        sourcePath := filepath.Join(os.TempDir(), "fake.ovf")
        expected := sourcePath

	vb := DefaultVagrantBox { sourcePath: sourcePath } 
        actual, err := vb.Expand(".ovf")

        assert := assert.New(t) 

        assert.Equal(expected, actual)
        assert.Nil(err)
        assert.Nil(vb.tempDir)
        assert.Empty(vb.vmPath)
}

func TestVagrantBox_ReturnErrorIfTempDirCannotBeCreated(t *testing.T) {
        sourcePath := filepath.Join(os.TempDir(), "fake.box")

        expected := sourcePath 
        expectedErrorMessage := "expected-create-dir-failure"

	vb := DefaultVagrantBox { sourcePath: sourcePath, tempDir: &MockVagrantBoxTempDir{ errorMessage: expectedErrorMessage } } 

        actual, err := vb.Expand(".ovf")

        assert := assert.New(t) 
        assert.Equal(expected, actual)
        assert.NotNil(err)
        assert.Equal(expectedErrorMessage, err.Error())
        assert.Equal("don't care", vb.tempDir.Path())
        assert.Empty(vb.vmPath)
}

func TestVagrantBox_ReturnErrorIfExtractFails(t *testing.T) {
        sourcePath := filepath.Join(os.TempDir(), "fake.box")

        expected := sourcePath
        expectedErrorMessage := "expected-targz-failure"

	vb := DefaultVagrantBox { 
                sourcePath: sourcePath,
                tempDir: &DefaultVagrantBoxTempDir{},  
                expander: &MockVagrantBoxExpander{ errorMessage: expectedErrorMessage },  
        }

        actual, err := vb.Expand(".ovf")

        assert := assert.New(t) 
        assert.Equal(expected, actual)
        assert.NotNil(err)
        assert.Equal(expectedErrorMessage, err.Error())
        assert.NotEmpty(vb.tempDir.Path())
        assert.Empty(vb.vmPath)
}

func TestVagrantBox_ReturnExpandedPathAndErrorIfExtractPassesButNoSuffixFound(t *testing.T) {
        sourcePath := filepath.Join(os.TempDir(), "fake.box")

	vb := DefaultVagrantBox { 
                sourcePath: sourcePath, 
                tempDir: &DefaultVagrantBoxTempDir{},  
                expander: &MockVagrantBoxExpander{},  
        }

        actual, err := vb.Expand(".ovf")
        expected := vb.tempDir.Path()

        assert := assert.New(t) 
        assert.Equal(expected, actual)
        assert.NotNil(err)

        expectedMessage := ".ovf not found in " + expected
        assert.Equal(expectedMessage, err.Error())
        assert.NotEmpty(vb.tempDir.Path())
        assert.Empty(vb.vmPath)
}

func TestVagrantBox_ReturnExpandedVMPathAndErrorIfExtractPassesAndSuffixFound(t *testing.T) {
        sourcePath := filepath.Join(os.TempDir(), "fake.box")

	vb := DefaultVagrantBox{
                sourcePath: sourcePath, 
                tempDir: &MockVagrantBoxTempDir{ fileInfo: &MockFileInfo { fileName: "expected.ovf" } },  
                expander: &MockVagrantBoxExpander{},
        }

        actual, err := vb.Expand(".ovf")

        assert := assert.New(t) 
        assert.Nil(err)
        assert.NotEmpty(vb.tempDir.Path())

        expected := filepath.Join(vb.tempDir.Path(), "expected.ovf")

        assert.Equal(expected, actual)
        assert.Equal(expected, vb.vmPath)
}
