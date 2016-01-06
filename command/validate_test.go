package command

import (
	"path/filepath"
	"testing"
)

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
