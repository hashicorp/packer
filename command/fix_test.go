package command

import (
	"github.com/hashicorp/packer/fix"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func TestFix(t *testing.T) {
	s := &strings.Builder{}
	ui := &packer.BasicUi{
		Writer: s,
	}
	c := &FixCommand{
		Meta: testMeta(t),
	}

	c.Ui = ui

	args := []string{filepath.Join(testFixture("fix"), "template.json")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	expected := `{
  "builders": [
    {
      "type": "dummy"
    }
  ],
  "push": {
    "name": "foo/bar"
  }
}`
	assert.Equal(t, expected, strings.TrimSpace(s.String()))
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

func TestFix_allFixersEnabled(t *testing.T) {
	f := fix.Fixers
	o := fix.FixerOrder

	if len(f) != len(o) {
		t.Fatalf("Fixers length (%d) does not match FixerOrder length (%d)", len(f), len(o))
	}

	for fixer, _ := range f {
		found := false

		for _, orderedFixer := range o {
			if orderedFixer == fixer {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Did not find Fixer %s in FixerOrder", fixer)
		}
	}
}
