// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

func TestFix(t *testing.T) {
	s := &strings.Builder{}
	ui := &packersdk.BasicUi{
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
