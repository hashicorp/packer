// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

const fixturesDir = "./test-fixtures"

func fatalCommand(t *testing.T, m Meta) {
	ui := m.Ui.(*packersdk.BasicUi)
	out := ui.Writer.(*bytes.Buffer)
	err := ui.ErrorWriter.(*bytes.Buffer)
	t.Fatalf(
		"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
		out.String(),
		err.String())
}

func testFixtureContent(n ...string) string {
	path := filepath.Join(append([]string{fixturesDir}, n...)...)
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func testFixture(n ...string) string {
	paths := []string{fixturesDir}
	paths = append(paths, n...)
	return filepath.Join(paths...)
}

func testMeta(t *testing.T) Meta {
	var out, err bytes.Buffer

	return Meta{
		CoreConfig: packer.TestCoreConfig(t),
		Ui: &packersdk.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}
