// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"fmt"
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

const (
	unformattedHCL = `
ami_filter_name ="amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners =[ "137112412989" ]

`
	formattedHCL = `
ami_filter_name   = "amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners = ["137112412989"]

`
)

func TestFmt_Recursive(t *testing.T) {

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
				"foo/bar/baz.pkr.hcl":         unformattedHCL,
				"foo/bar/baz/woo.pkrvars.hcl": unformattedHCL,
				"potato":                      unformattedHCL,
				"foo/bar/potato":              unformattedHCL,
				"bar.pkr.hcl":                 unformattedHCL,
				"-":                           unformattedHCL,
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo/bar/baz.pkr.hcl":         formattedHCL,
					"foo/bar/baz/woo.pkrvars.hcl": formattedHCL,
					"potato":                      unformattedHCL,
					"foo/bar/potato":              unformattedHCL,
					"bar.pkr.hcl":                 formattedHCL,
					"-":                           unformattedHCL,
				}},
		},
		{
			name:       "nested no recursive format",
			formatArgs: []string{},
			alreadyPresentContent: map[string]string{
				"foo/bar/baz.pkr.hcl":         unformattedHCL,
				"foo/bar/baz/woo.pkrvars.hcl": unformattedHCL,
				"bar.pkr.hcl":                 unformattedHCL,
				"-":                           unformattedHCL,
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo/bar/baz.pkr.hcl":         unformattedHCL,
					"foo/bar/baz/woo.pkrvars.hcl": unformattedHCL,
					"bar.pkr.hcl":                 formattedHCL,
					"-":                           unformattedHCL,
				}},
		},
	}

	c := &FormatCommand{
		Meta: testMeta(t),
	}

	testDir := "test-fixtures/fmt"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDirectory := mustString(os.MkdirTemp(testDir, "test-dir-*"))
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

func Test_fmt_pipe(t *testing.T) {

	tc := []struct {
		piped    string
		command  []string
		env      []string
		expected string
	}{
		{unformattedHCL, []string{"fmt", "-"}, nil, formattedHCL},
		{formattedHCL, []string{"fmt", "-"}, nil, formattedHCL},
	}

	for _, tc := range tc {
		t.Run(fmt.Sprintf("echo %q | packer %s", tc.piped, tc.command), func(t *testing.T) {
			p := helperCommand(t, tc.command...)
			p.Stdin = strings.NewReader(tc.piped)
			p.Env = append(p.Env, tc.env...)
			fmt.Println(fmt.Sprintf("Path: %s", p.Path))
			bs, err := p.Output()
			if err != nil {
				t.Fatalf("Error occurred running command %v: %s", err, bs)
			}
			if diff := cmp.Diff(tc.expected, string(bs)); diff != "" {
				t.Fatalf("Error in diff: %s", diff)
			}
		})
	}
}

const malformedTemplate = "test-fixtures/fmt_errs/malformed.pkr.hcl"

func TestFmtParseError(t *testing.T) {
	p := helperCommand(t, "fmt", malformedTemplate)
	outs, err := p.CombinedOutput()
	if err == nil {
		t.Errorf("Expected failure to format file, but command did not fail")
	}
	strLogs := string(outs)

	if !strings.Contains(strLogs, "An argument or block definition is required here.") {
		t.Errorf("Expected some diags about parse error, found none")
	}
}
