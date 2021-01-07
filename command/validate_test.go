package command

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestValidateCommand(t *testing.T) {
	tt := []struct {
		path     string
		exitCode int
	}{
		{path: filepath.Join(testFixture("validate"), "build.json")},
		{path: filepath.Join(testFixture("validate"), "build.pkr.hcl")},
		{path: filepath.Join(testFixture("validate"), "build_with_vars.pkr.hcl")},
		{path: filepath.Join(testFixture("validate-invalid"), "bad_provisioner.json"), exitCode: 1},
		{path: filepath.Join(testFixture("validate-invalid"), "missing_build_block.pkr.hcl"), exitCode: 1},
		{path: filepath.Join(testFixture("validate"), "null_var.json"), exitCode: 1},
		{path: filepath.Join(testFixture("validate"), "var_foo_with_no_default.pkr.hcl"), exitCode: 1},

		// wrong version fails
		{path: filepath.Join(testFixture("version_req", "base_failure")), exitCode: 1},
		{path: filepath.Join(testFixture("version_req", "base_success")), exitCode: 0},

		// wrong version field
		{path: filepath.Join(testFixture("version_req", "wrong_field_name")), exitCode: 1},

		// wrong packer block
		{path: filepath.Join(testFixture("validate", "invalid_packer_block.pkr.hcl")), exitCode: 1},
	}

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			c := &ValidateCommand{
				Meta: testMetaFile(t),
			}
			tc := tc
			args := []string{tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
	}
}

func TestValidateCommand_SkipDatasourceExecution(t *testing.T) {
	datasourceMock := &packersdk.MockDatasource{}
	meta := testMetaFile(t)
	meta.CoreConfig.Components.DatasourceStore = packersdk.MapOfDatasource{
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
				Meta: testMetaFile(t),
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
		Meta: testMetaFile(t),
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
		Meta: testMetaFile(t),
	}
	args := []string{
		filepath.Join(testFixture("validate"), "template.json"),
	}

	// This should fail with an invalid configuration version
	c.CoreConfig.Version = "100.0.0"
	if code := c.Run(args); code != 1 {
		t.Errorf("Expected exit code 1")
	}

	stdout, stderr := outputCommand(t, c.Meta)
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
		Meta: testMetaFile(t),
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
