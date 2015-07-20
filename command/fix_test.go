package command

import (
	"path/filepath"
	"testing"
)

func TestFix_noArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run(nil)
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}

func TestFix_multiArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run([]string{"one", "two"})
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}

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
