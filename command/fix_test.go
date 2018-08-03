package command

import (
	"path/filepath"
	"testing"
)

func TestFix(t *testing.T) {
	c := &FixCommand{
		Meta: testMeta(t),
	}

	args := []string{filepath.Join(testFixture("fix"), "template.json")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
}

func TestFix_invalidTemplate(t *testing.T) {
	c := &FixCommand{
		Meta: testMeta(t),
	}

	args := []string{filepath.Join(testFixture("fix-invalid"), "template.json")}
	if code := c.Run(args); code != 1 {
		fatalCommand(t, c.Meta)
	}
}

func TestFix_invalidTemplateDisableValidation(t *testing.T) {
	c := &FixCommand{
		Meta: testMeta(t),
	}

	args := []string{
		"-validate=false",
		filepath.Join(testFixture("fix-invalid"), "template.json"),
	}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
}
