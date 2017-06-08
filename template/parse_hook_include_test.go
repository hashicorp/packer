package template

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseHookInclude(t *testing.T) {
	cases := []struct {
		File   string
		Result *Template
		Err    bool
	}{
		{
			"parse-hook-include-main.json",
			&Template{
				Variables: map[string]*Variable{
					"builder_type": {
						Default: "builder-one",
					},
					"with_gui": {
						Default: "yes",
					},
				},
			},
			false,
		},
		{
			"parse-hook-include-main-override.json",
			&Template{
				Variables: map[string]*Variable{
					"builder_type": {
						Default: "builder-overriden",
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		path, _ := filepath.Abs(fixtureDir(tc.File))
		tpl, err := ParseFile(fixtureDir(tc.File))
		if (err != nil) != tc.Err {
			t.Fatalf("err: %#v", err)
		}

		if tc.Result != nil {
			tc.Result.Path = path
		}
		if tpl != nil {
			tpl.RawContents = nil
		}
		if !reflect.DeepEqual(tpl, tc.Result) {
			t.Fatalf("bad: %s\n\n%#v\n\n%#v", tc.File, tpl, tc.Result)
		}
	}
}

func TestParseHookInclude_bad(t *testing.T) {
	cases := []struct {
		File     string
		Expected string
	}{
		{"parse-hook-include-main-invalid-value.json", "neither a string nor an array"},
		{"parse-hook-include-main-invalid-value-in-array.json", "not a string"},
	}
	for _, tc := range cases {
		_, err := ParseFile(fixtureDir(tc.File))
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), tc.Expected) {
			t.Fatalf("file: %s\nExpected: %s\n%s\n", tc.File, tc.Expected, err.Error())
		}
	}
}
