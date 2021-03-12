package command

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

func TestFmt(t *testing.T) {
	s := &strings.Builder{}
	ui := &packersdk.BasicUi{
		Writer: s,
	}
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	c.Ui = ui

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "formatted.pkr.hcl")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	expected := ""
	assert.Equal(t, expected, strings.TrimSpace(s.String()))
}

func TestFmt_unformattedPKRVarsTemplate(t *testing.T) {
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "unformatted.pkrvars.hcl")}
	if code := c.Run(args); code != 3 {
		fatalCommand(t, c.Meta)
	}
}

func TestFmt_unfomattedTemlateDirectory(t *testing.T) {
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "")}

	if code := c.Run(args); code != 3 {
		fatalCommand(t, c.Meta)
	}
}

func TestFmt_Recursive(t *testing.T) {
	unformattedData := `
ami_filter_name ="amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners =[ "137112412989" ]

`

	formattedData := `
ami_filter_name   = "amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners = ["137112412989"]

`

	tests := []struct {
		name                  string
		formatArgs            []string // arguments passed to format
		alreadyPresentContent map[string]string
		fileCheck
	}{
		{
			name:       "nested formats recursively",
			formatArgs: []string{"-recursive=true"},
			alreadyPresentContent: map[string]string{
				"foo/bar/baz.pkr.hcl":         unformattedData,
				"foo/bar/baz/woo.pkrvars.hcl": unformattedData,
				"potato":                      unformattedData,
				"foo/bar/potato":              unformattedData,
				"bar.pkr.hcl":                 unformattedData,
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo/bar/baz.pkr.hcl":         formattedData,
					"foo/bar/baz/woo.pkrvars.hcl": formattedData,
					"potato":                      unformattedData,
					"foo/bar/potato":              unformattedData,
					"bar.pkr.hcl":                 formattedData,
				}},
		},
		{
			name:       "nested no recursive format",
			formatArgs: []string{},
			alreadyPresentContent: map[string]string{
				"foo/bar/baz.pkr.hcl":         unformattedData,
				"foo/bar/baz/woo.pkrvars.hcl": unformattedData,
				"bar.pkr.hcl":                 unformattedData,
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo/bar/baz.pkr.hcl":         unformattedData,
					"foo/bar/baz/woo.pkrvars.hcl": unformattedData,
					"bar.pkr.hcl":                 formattedData,
				}},
		},
	}

	c := &FormatCommand{
		Meta: testMeta(t),
	}

	testDir := "test-fixtures/fmt"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDirectory := mustString(ioutil.TempDir(testDir, "test-dir-*"))
			defer os.RemoveAll(tempDirectory)

			createFiles(tempDirectory, tt.alreadyPresentContent)

			testArgs := append(tt.formatArgs, tempDirectory)
			if code := c.Run(testArgs); code != 0 {
				ui := c.Meta.Ui.(*packersdk.BasicUi)
				out := ui.Writer.(*bytes.Buffer)
				err := ui.ErrorWriter.(*bytes.Buffer)
				t.Fatalf(
					"Bad exit code for test case: %s.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
					tt.name,
					out.String(),
					err.String())
			}

			tt.fileCheck.verify(t, tempDirectory)
		})
	}
}
