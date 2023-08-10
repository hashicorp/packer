// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

func TestValidateCommand(t *testing.T) {
	tt := []struct {
		path      string
		exitCode  int
		extraArgs []string
	}{
		{path: filepath.Join(testFixture("validate"), "build.json")},
		{path: filepath.Join(testFixture("validate"), "build.pkr.hcl")},
		{path: filepath.Join(testFixture("validate"), "build_with_vars.pkr.hcl")},
		{path: filepath.Join(testFixture("validate-invalid"), "bad_provisioner.json"), exitCode: 1},
		{path: filepath.Join(testFixture("validate-invalid"), "missing_build_block.pkr.hcl"), exitCode: 1},
		{path: filepath.Join(testFixture("validate"), "null_var.json"), exitCode: 1},
		{path: filepath.Join(testFixture("validate"), "var_foo_with_no_default.pkr.hcl"), exitCode: 1},

		{path: testFixture("hcl", "validation", "wrong_pause_before.pkr.hcl"), exitCode: 1},

		// wrong version fails
		{path: filepath.Join(testFixture("version_req", "base_failure")), exitCode: 1},
		{path: filepath.Join(testFixture("version_req", "base_success")), exitCode: 0},

		// wrong version field
		{path: filepath.Join(testFixture("version_req", "wrong_field_name")), exitCode: 1},

		// wrong packer block
		{path: filepath.Join(testFixture("validate", "invalid_packer_block.pkr.hcl")), exitCode: 1},

		// Should return multiple errors,
		{path: filepath.Join(testFixture("validate", "circular_error.pkr.hcl")), exitCode: 1},

		// datasource could be unknown at that moment
		{path: filepath.Join(testFixture("hcl", "data-source-validation.pkr.hcl")), exitCode: 0},

		// datasource unknown at validation-time without datasource evaluation -> fail on provisioner
		{path: filepath.Join(testFixture("hcl", "local-ds-validate.pkr.hcl")), exitCode: 1},
		// datasource unknown at validation-time with datasource evaluation -> success
		{path: filepath.Join(testFixture("hcl", "local-ds-validate.pkr.hcl")), exitCode: 0, extraArgs: []string{"--evaluate-datasources"}},
	}

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			tc := tc
			args := tc.extraArgs
			args = append(args, tc.path)
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
	}
}

func TestValidateCommand_SkipDatasourceExecution(t *testing.T) {
	datasourceMock := &packersdk.MockDatasource{}
	meta := TestMetaFile(t)
	meta.CoreConfig.Components.PluginConfig.DataSources = packer.MapOfDatasource{
		"mock": func() (packersdk.Datasource, error) {
			return datasourceMock, nil
		},
	}
	c := &ValidateCommand{
		Meta: meta,
	}
	args := []string{filepath.Join(testFixture("validate"), "datasource.pkr.hcl")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	if datasourceMock.ExecuteCalled {
		t.Fatalf("Datasource should not be executed on validation")
	}
	if !datasourceMock.OutputSpecCalled {
		t.Fatalf("Datasource OutPutSpec should be called on validation")
	}
}

func TestValidateCommand_SyntaxOnly(t *testing.T) {
	tt := []struct {
		path     string
		exitCode int
	}{
		{path: filepath.Join(testFixture("validate"), "build.json")},
		{path: filepath.Join(testFixture("validate"), "build.pkr.hcl")},
		{path: filepath.Join(testFixture("validate"), "build_with_vars.pkr.hcl")},
		{path: filepath.Join(testFixture("validate-invalid"), "bad_provisioner.json")},
		{path: filepath.Join(testFixture("validate-invalid"), "missing_build_block.pkr.hcl")},
		{path: filepath.Join(testFixture("validate-invalid"), "broken.json"), exitCode: 1},
		{path: filepath.Join(testFixture("validate"), "null_var.json")},
		{path: filepath.Join(testFixture("validate"), "var_foo_with_no_default.pkr.hcl")},
	}

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			c.CoreConfig.Version = "102.0.0"
			tc := tc
			args := []string{"-syntax-only", tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
	}
}

func TestValidateCommandOKVersion(t *testing.T) {
	c := &ValidateCommand{
		Meta: TestMetaFile(t),
	}
	args := []string{
		filepath.Join(testFixture("validate"), "template.json"),
	}

	// This should pass with a valid configuration version
	c.CoreConfig.Version = "102.0.0"
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
}

func TestValidateCommandBadVersion(t *testing.T) {
	c := &ValidateCommand{
		Meta: TestMetaFile(t),
	}
	args := []string{
		filepath.Join(testFixture("validate"), "template.json"),
	}

	// This should fail with an invalid configuration version
	c.CoreConfig.Version = "100.0.0"
	if code := c.Run(args); code != 1 {
		t.Errorf("Expected exit code 1")
	}

	stdout, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
	expected := `Error: 

This template requires Packer version 101.0.0 or higher; using 100.0.0


`

	if diff := cmp.Diff(expected, stderr); diff != "" {
		t.Errorf("Unexpected output: %s", diff)
	}
	t.Log(stdout)
}

func TestValidateCommandExcept(t *testing.T) {
	tt := []struct {
		name     string
		args     []string
		exitCode int
	}{
		{
			name: "JSON: validate except build and post-processor",
			args: []string{
				"-except=vanilla,pear",
				filepath.Join(testFixture("validate"), "validate_except.json"),
			},
		},
		{
			name: "JSON: fail validate except build and post-processor",
			args: []string{
				"-except=chocolate,apple",
				filepath.Join(testFixture("validate"), "validate_except.json"),
			},
			exitCode: 1,
		},
		{
			name: "HCL2: validate except build and post-processor",
			args: []string{
				"-except=file.vanilla,pear",
				filepath.Join(testFixture("validate"), "validate_except.pkr.hcl"),
			},
		},
		{
			name: "HCL2: fail validation except build and post-processor",
			args: []string{
				"-except=file.chocolate,apple",
				filepath.Join(testFixture("validate"), "validate_except.pkr.hcl"),
			},
			exitCode: 1,
		},
	}

	c := &ValidateCommand{
		Meta: TestMetaFile(t),
	}
	c.CoreConfig.Version = "102.0.0"

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer cleanup()

			tc := tc
			if code := c.Run(tc.args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
	}
}

func TestValidateCommand_VarFiles(t *testing.T) {
	tt := []struct {
		name     string
		path     string
		varfile  string
		exitCode int
	}{
		{name: "with basic HCL var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkrvars.hcl"),
			exitCode: 0,
		},
		{name: "with unused variable in var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.pkrvars.hcl"),
			exitCode: 0,
		},
		{name: "with unused variable in JSON var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.json"),
			exitCode: 0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			tc := tc
			args := []string{"-var-file", tc.varfile, tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
	}
}

func TestValidateCommand_VarFilesWarnOnUndeclared(t *testing.T) {
	tt := []struct {
		name     string
		path     string
		varfile  string
		exitCode int
	}{
		{name: "default warning with unused variable in HCL var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.pkrvars.hcl"),
			exitCode: 0,
		},
		{name: "default warning with unused variable in JSON var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.json"),
			exitCode: 0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			tc := tc
			args := []string{"-var-file", tc.varfile, tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}

			stdout, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			expected := `Warning: Undefined variable

The variable "unused" was set but was not declared as an input variable.
To declare variable "unused" place this block in one of your .pkr.hcl files,
such as variables.pkr.hcl

variable "unused" {
  type    = string
  default = null
}


The configuration is valid.
`
			if diff := cmp.Diff(expected, stdout); diff != "" {
				t.Errorf("Unexpected output: %s", diff)
			}
			t.Log(stderr)
		})
	}
}

func TestValidateCommand_VarFilesDisableWarnOnUndeclared(t *testing.T) {
	tt := []struct {
		name     string
		path     string
		varfile  string
		exitCode int
	}{
		{name: "no-warn-undeclared-var with unused variable in HCL var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.pkrvars.hcl"),
			exitCode: 0,
		},
		{name: "no-warn-undeclared-var with unused variable in JSON var-file definition",
			path:     filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "basic.pkr.hcl"),
			varfile:  filepath.Join(testFixture(filepath.Join("validate", "var-file-tests")), "undeclared.json"),
			exitCode: 0,
		},
	}
	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			tc := tc
			args := []string{"-no-warn-undeclared-var", "-var-file", tc.varfile, tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}

			stdout, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			expected := `The configuration is valid.
`
			if diff := cmp.Diff(expected, stdout); diff != "" {
				t.Errorf("Unexpected output: %s", diff)
			}
			t.Log(stderr)
		})
	}
}

func TestValidateCommand_ShowLineNumForMissing(t *testing.T) {
	tt := []struct {
		path      string
		exitCode  int
		extraArgs []string
	}{
		{path: filepath.Join(testFixture("validate-invalid"), "missing_build_block.pkr.hcl"), exitCode: 1},
	}

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: TestMetaFile(t),
			}
			tc := tc
			args := tc.extraArgs
			args = append(args, tc.path)
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}

			stdout, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			expected := fmt.Sprintf(`Error: Unknown source file.cho

  on %s line 6:
  (source code not available)

Known: [file.chocolate]


`, tc.path)
			if diff := cmp.Diff(expected, stderr); diff != "" {
				t.Errorf("Unexpected output: %s", diff)
			}
			t.Log(stdout)
		})
	}
}
