package template

import (
	"os"
	"path/filepath"
	"testing"
)

const FixturesDir = "./test-fixtures"

// fixtureDir returns the path to a test fixtures directory
func fixtureDir(n string) string {
	return filepath.Join(FixturesDir, n)
}

func TestTemplateValidate(t *testing.T) {
	cases := []struct {
		File string
		Err  bool
	}{
		{
			"validate-no-builders.json",
			true,
		},

		{
			"validate-bad-override.json",
			true,
		},

		{
			"validate-good-override.json",
			false,
		},

		{
			"validate-bad-prov-only.json",
			true,
		},

		{
			"validate-good-prov-only.json",
			false,
		},

		{
			"validate-bad-prov-except.json",
			true,
		},

		{
			"validate-good-prov-except.json",
			false,
		},
	}

	for _, tc := range cases {
		f, err := os.Open(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		tpl, err := Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		err = tpl.Validate()
		if (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
	}
}
