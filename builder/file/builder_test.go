package file

import (
	"fmt"
	"io/ioutil"
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packer.Builder = new(Builder)
}

func TestBuilderFileAcc_content(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: fileContentTest,
		Check:    checkContent,
	})
}

func TestBuilderFileAcc_copy(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: fileCopyTest,
		Check:    checkCopy,
	})
}

func checkContent(artifacts []packersdk.Artifact) error {
	content, err := ioutil.ReadFile("contentTest.txt")
	if err != nil {
		return err
	}
	contentString := string(content)
	if contentString != "hello world!" {
		return fmt.Errorf("Unexpected file contents: %s", contentString)
	}
	return nil
}

func checkCopy(artifacts []packersdk.Artifact) error {
	content, err := ioutil.ReadFile("copyTest.txt")
	if err != nil {
		return err
	}
	contentString := string(content)
	if contentString != "Hello world.\n" {
		return fmt.Errorf("Unexpected file contents: %s", contentString)
	}
	return nil
}

const fileContentTest = `
{
    "builders": [
        {
            "type":"test",
            "target":"contentTest.txt",
            "content":"hello world!"
        }
    ]
}
`

const fileCopyTest = `
{
    "builders": [
        {
            "type":"test",
            "target":"copyTest.txt",
            "source":"test-fixtures/artifact.txt"
        }
    ]
}
`
