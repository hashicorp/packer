package command

import (
	"path/filepath"
	"testing"
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
	}

	c := &ValidateCommand{
		Meta: testMetaFile(t),
	}

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
			tc := tc
			args := []string{tc.path}
			if code := c.Run(args); code != tc.exitCode {
				fatalCommand(t, c.Meta)
			}
		})
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
	}

	c := &ValidateCommand{
		Meta: testMetaFile(t),
	}
	c.CoreConfig.Version = "102.0.0"

	for _, tc := range tt {
		t.Run(tc.path, func(t *testing.T) {
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
	expected := `Error initializing core: This template requires Packer version 101.0.0 or higher; using 100.0.0
`
	if stderr != expected {
		t.Fatalf("Expected:\n%s\nFound:\n%s\n", expected, stderr)
	}
	t.Log(stdout)
}
