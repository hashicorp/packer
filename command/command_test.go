package command

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/mitchellh/packer/packer"
)

const fixturesDir = "./test-fixtures"

func fatalCommand(t *testing.T, m Meta) {
	ui := m.Ui.(*packer.BasicUi)
	out := ui.Writer.(*bytes.Buffer)
	err := ui.ErrorWriter.(*bytes.Buffer)
	t.Fatalf(
		"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
		out.String(),
		err.String())
}

func outputCommand(t *testing.T, m Meta) (string, string) {
	ui := m.Ui.(*packer.BasicUi)
	out := ui.Writer.(*bytes.Buffer)
	err := ui.ErrorWriter.(*bytes.Buffer)
	return out.String(), err.String()
}

func testFixture(n string) string {
	return filepath.Join(fixturesDir, n)
}

func testMeta(t *testing.T) Meta {
	var out, err bytes.Buffer

	return Meta{
		CoreConfig: packer.TestCoreConfig(t),
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}
