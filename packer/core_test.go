package packer

import (
	"os"
	"testing"

	"github.com/mitchellh/packer/template"
)

func TestCoreValidate(t *testing.T) {
	cases := []struct {
		File string
		Vars map[string]string
		Err  bool
	}{
		{
			"validate-dup-builder.json",
			nil,
			true,
		},

		// Required variable not set
		{
			"validate-req-variable.json",
			nil,
			true,
		},

		{
			"validate-req-variable.json",
			map[string]string{"foo": "bar"},
			false,
		},
	}

	for _, tc := range cases {
		f, err := os.Open(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		tpl, err := template.Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		core, err := NewCore(&CoreConfig{
			Template:  tpl,
			Variables: tc.Vars,
		})
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		if err := core.Validate(); (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
	}
}
