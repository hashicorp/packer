package command

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	unformattedData := `ami_filter_name ="amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners =[ "137112412989" ]

`

	formattedData := `ami_filter_name   = "amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners = ["137112412989"]

`

	tests := []struct {
		name                  string
		formatArgs            []string // arguments passed to format
		alreadyPresentContent map[string]string
		expectedContent       map[string]string
	}{
		{
			name:       "nested formats recursively",
			formatArgs: []string{"-recursive=true"},
			alreadyPresentContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                unformattedData,
			},
			expectedContent: map[string]string{
				"foo/bar/baz":     formattedData,
				"foo/bar/baz/woo": formattedData,
				"":                formattedData,
			},
		},
		{
			name:       "nested no recursive format",
			formatArgs: []string{},
			alreadyPresentContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                unformattedData,
			},
			expectedContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                formattedData,
			},
		},
	}

	c := &FormatCommand{
		Meta: testMeta(t),
	}

	testDir := "test-fixtures/fmt"

	for _, tt := range tests {
		tempFileNames := make(map[string]string)

		tempDirectory, err := ioutil.TempDir(testDir, "test-dir-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir for test case: %s, error: %v", tt.name, err)
		}
		defer os.RemoveAll(tempDirectory)

		for subDir, content := range tt.alreadyPresentContent {
			dir := filepath.Join(tempDirectory, subDir)
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				t.Fatalf("Failed to create directory for test case: %s, error: %v", tt.name, err)
			}

			tempFile, err := ioutil.TempFile(dir, "*.pkrvars.hcl")
			if err != nil {
				t.Fatalf("Failed to create temp file for test case: %s, error: %v", tt.name, err)
			}

			_, err = tempFile.Write([]byte(content))
			if err != nil {
				t.Fatalf("Failed to write temp file for test case: %s, error: %v", tt.name, err)
			}
			tempFileNames[subDir] = tempFile.Name()
			tempFile.Close()
		}

		testArgs := append(tt.formatArgs, tempDirectory)
		if code := c.Run(testArgs); code != 0 {
			os.RemoveAll(tempDirectory)
			ui := c.Meta.Ui.(*packersdk.BasicUi)
			out := ui.Writer.(*bytes.Buffer)
			err := ui.ErrorWriter.(*bytes.Buffer)
			t.Fatalf(
				"Bad exit code for test case: %s.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
				tt.name,
				out.String(),
				err.String())
		}

		for expectedPath, expectedContent := range tt.expectedContent {
			b, err := ioutil.ReadFile(tempFileNames[expectedPath])
			if err != nil {
				t.Fatalf("ReadFile failed for test case: %s, error : %v", tt.name, err)
			}
			got := string(b)
			if diff := cmp.Diff(got, expectedContent); diff != "" {
				t.Errorf(
					"format dir, unexpected result for test case: %s, path: %s,  Expected: %s, Got: %s",
					tt.name,
					expectedPath,
					expectedContent,
					got)
			}
		}
	}

}
