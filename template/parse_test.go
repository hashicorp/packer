package template

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		File   string
		Result *Template
		Err    bool
	}{
		{
			"parse-basic.json",
			&Template{
				Builders: map[string]*Builder{
					"something": &Builder{
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-builder-no-type.json",
			nil,
			true,
		},
		{
			"parse-builder-repeat.json",
			nil,
			true,
		},
	}

	for _, tc := range cases {
		f, err := os.Open(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		tpl, err := Parse(f)
		f.Close()
		if (err != nil) != tc.Err {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(tpl, tc.Result) {
			t.Fatalf("bad: %#v", tpl)
		}
	}
}
