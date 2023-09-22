// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package file

import (
	"fmt"
	"os"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	builderT "github.com/hashicorp/packer/acctest"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packersdk.Builder = new(Builder)
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
	content, err := os.ReadFile("contentTest.txt")
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
	content, err := os.ReadFile("copyTest.txt")
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
