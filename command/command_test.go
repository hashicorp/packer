// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
	"github.com/posener/complete"
	"github.com/stretchr/testify/require"
)

const fixturesDir = "./test-fixtures"

func TestAutocompleteArgs(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("template.pkr.hcl", nil, 0600))
	require.NoError(t, os.WriteFile("template.json", nil, 0600))
	require.NoError(t, os.Mkdir("templates", 0700))

	for name, predictor := range map[string]complete.Predictor{
		"build":              (&BuildCommand{}).AutocompleteArgs(),
		"console":            (&ConsoleCommand{}).AutocompleteArgs(),
		"fix":                (&FixCommand{}).AutocompleteArgs(),
		"fmt":                (&FormatCommand{}).AutocompleteArgs(),
		"hcl2_upgrade":       (&HCL2UpgradeCommand{}).AutocompleteArgs(),
		"init":               (&InitCommand{}).AutocompleteArgs(),
		"inspect":            (&InspectCommand{}).AutocompleteArgs(),
		"validate":           (&ValidateCommand{}).AutocompleteArgs(),
		"verify-attestation": (&VerifyAttestationCommand{}).AutocompleteArgs(),
	} {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, predictor)
			got := predictor.Predict(complete.Args{Last: "templ"})
			require.ElementsMatch(t, []string{"template.pkr.hcl", "template.json", "templates/"}, got)
		})
	}
}

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
