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

	testFileName := "test.pkrvars.hcl"

	for _, tt := range tests {
		topDir, err := ioutil.TempDir("test-fixtures/fmt", "temp-dir")
		if err != nil {
			t.Fatalf("Failed to create TopDir for test case: %s, error: %v", tt.name, err)
		}
		defer os.Remove(topDir)

		for testDir, content := range tt.alreadyPresentContent {
			dir := filepath.Join(topDir, testDir)
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf(
					"Failed to create subDir: %s\n\n, for test case: %s\n\n, error: %v",
					testDir,
					tt.name,
					err)
			}

			file, err := os.Create(filepath.Join(dir, testFileName))
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to create testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}

			_, err = file.Write([]byte(content))
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to write to testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}

			err = file.Close()
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to close testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}
		}

		testArgs := append(tt.formatArgs, topDir)
		if code := c.Run(testArgs); code != 0 {
			os.RemoveAll(topDir)
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
			b, err := ioutil.ReadFile(filepath.Join(topDir, expectedPath, testFileName))
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("ReadFile failed for test case: %s, error : %v", tt.name, err)
			}
			got := string(b)
			if diff := cmp.Diff(got, expectedContent); diff != "" {
				os.RemoveAll(topDir)
				t.Errorf(
					"format dir, unexpected result for test case: %s, path: %s,  Expected: %s, Got: %s",
					tt.name,
					expectedPath,
					expectedContent,
					got)
			}
		}

		err = os.RemoveAll(topDir)
		if err != nil {
			t.Errorf(
				"Failed to delete top level test directory for test case: %s, please clean before another test run. Error: %s",
				tt.name,
				err)
		}
	}

}
