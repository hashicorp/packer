package command

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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

func outputCommand(t *testing.T, m Meta) (string, string) {
	ui := m.Ui.(*packersdk.BasicUi)
	out := ui.Writer.(*bytes.Buffer)
	err := ui.ErrorWriter.(*bytes.Buffer)
	return out.String(), err.String()
}

func testFixtureContent(n ...string) string {
	path := filepath.Join(append([]string{fixturesDir}, n...)...)
	b, err := ioutil.ReadFile(path)
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
