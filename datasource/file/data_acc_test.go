// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package file

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestFileDataSource(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		createOutput bool
		expectError  bool
		expectOutput string
	}{
		{
			"Success - write empty file",
			basicEmptyFileWrite,
			false,
			false,
			"",
		},
		{
			"Fail - write empty file, pre-existing output",
			basicEmptyFileWrite,
			true,
			true,
			"",
		},
		{
			"Success - write empty file, pre-existing output",
			basicEmptyFileWriteForce,
			true,
			false,
			"",
		},
		{
			"Success - write template to output",
			basicFileWithTemplateContents,
			false,
			false,
			"contents are 12345\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCase := &acctest.PluginTestCase{
				Name: tt.name,
				Setup: func() error {
					return nil
				},
				Teardown: func() error {
					return nil
				},
				Template: tt.template,
				Type:     "http",
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 && !tt.expectError {
							return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
						}
						if tt.expectError && buildCommand.ProcessState.ExitCode() == 0 {
							return fmt.Errorf("Expected an error but succeeded.")
						}
					}

					if tt.expectError {
						return nil
					}

					outFile, err := os.ReadFile("output")
					if err != nil {
						return fmt.Errorf("failed to read output file: %s", err)
					}

					diff := cmp.Diff(string(outFile), tt.expectOutput)
					if diff != "" {
						return fmt.Errorf("diff found in output: %s", diff)
					}

					return nil
				},
			}

			os.RemoveAll("output")
			if tt.createOutput {
				err := os.WriteFile("output", []byte{}, 0644)
				if err != nil {
					t.Fatalf("failed to pre-create output file: %s", err)
				}
			}

			acctest.TestPlugin(t, testCase)

			os.RemoveAll("output")
		})
	}
}

var basicEmptyFileWrite string = `
source "null" "test" {
	communicator = "none"
}

data "file" "empty" {
	destination = "output"
}

build {
	sources = [
		"source.null.test"
	]

	provisioner "shell-local" {
		inline = [
			"set -ex",
			"test -f ${data.file.empty.path}",
		]
	}
}
`

var basicEmptyFileWriteForce string = `
source "null" "test" {
	communicator = "none"
}

data "file" "empty" {
	destination = "output"
	force = true
}

build {
	sources = [
		"source.null.test"
	]

	provisioner "shell-local" {
		inline = [
			"set -ex",
			"test -f ${data.file.empty.path}",
		]
	}
}
`

var basicFileWithTemplateContents string = `
source "null" "test" {
	communicator = "none"
}

data "file" "empty" {
	contents = templatefile("test-fixtures/template.pkrtpl.hcl", {
		"value" = "12345",
	})
	destination = "output"
}

build {
	sources = [
		"source.null.test"
	]

	provisioner "shell-local" {
		inline = [
			"set -ex",
			"test -f ${data.file.empty.path}",
		]
	}
}
`
