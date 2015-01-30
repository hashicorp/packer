package command

import (
	"path/filepath"
	"testing"

	"github.com/mitchellh/cli"
)

const fixturesDir = "./test-fixtures"

func fatalCommand(t *testing.T, m Meta) {
	ui := m.Ui.(*cli.MockUi)
	t.Fatalf(
		"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
		ui.OutputWriter.String(),
		ui.ErrorWriter.String())
}

func testFixture(n string) string {
	return filepath.Join(fixturesDir, n)
}

func testMeta(t *testing.T) Meta {
	return Meta{
		Ui: new(cli.MockUi),
	}
}
