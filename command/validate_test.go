package command

import (
	"path/filepath"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	c := &ValidateCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		filepath.Join(testFixture("validate"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	if !fileExists("chocolate.txt") {
		t.Error("Expected to find chocolate.txt")
	}
	if !fileExists("vanilla.txt") {
		t.Error("Expected to find vanilla.txt")
	}
	if fileExists("cherry.txt") {
		t.Error("Expected NOT to find cherry.txt")
	}
}
