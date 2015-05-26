package packer

import (
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/packer/template"
)

func TestCoreBuildNames(t *testing.T) {
	cases := []struct {
		File   string
		Vars   map[string]string
		Result []string
	}{
		{
			"build-names-basic.json",
			nil,
			[]string{"something"},
		},

		{
			"build-names-func.json",
			nil,
			[]string{"TUBES"},
		},
	}

	for _, tc := range cases {
		tpl, err := template.ParseFile(fixtureDir(tc.File))
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

		names := core.BuildNames()
		if !reflect.DeepEqual(names, tc.Result) {
			t.Fatalf("err: %s\n\n%#v", tc.File, names)
		}
	}
}

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

func testComponentFinder() *ComponentFinder {
	builderFactory := func(n string) (Builder, error) { return new(MockBuilder), nil }
	ppFactory := func(n string) (PostProcessor, error) { return new(TestPostProcessor), nil }
	provFactory := func(n string) (Provisioner, error) { return new(MockProvisioner), nil }
	return &ComponentFinder{
		Builder:       builderFactory,
		PostProcessor: ppFactory,
		Provisioner:   provFactory,
	}
}
