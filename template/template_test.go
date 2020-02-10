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
			"validate-good-prov-timeout.json",
			false,
		},

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

		{
			"validate-bad-pp-only.json",
			true,
		},

		{
			"validate-good-pp-only.json",
			false,
		},

		{
			"validate-bad-pp-except.json",
			true,
		},

		{
			"validate-good-pp-except.json",
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

func TestOnlyExceptSkip(t *testing.T) {
	cases := []struct {
		Only, Except []string
		Input        string
		Result       bool
	}{
		{
			[]string{"foo"},
			nil,
			"foo",
			false,
		},

		{
			nil,
			[]string{"foo"},
			"foo",
			true,
		},

		{
			nil,
			nil,
			"foo",
			false,
		},
	}

	for _, tc := range cases {
		oe := &OnlyExcept{
			Only:   tc.Only,
			Except: tc.Except,
		}

		actual := oe.Skip(tc.Input)
		if actual != tc.Result {
			t.Fatalf(
				"bad: %#v\n\n%#v\n\n%#v\n\n%#v",
				actual, tc.Only, tc.Except, tc.Input)
		}
	}
}

func TestPostProcessor_IsValidToBuilder(t *testing.T) {
	pp := &PostProcessor{
		Name: "vsphere",
	}

	// valid builder vmware-iso
	valid := pp.IsValidWithBuilder("vmware-iso")
	if !valid {
		t.Fatalf("vsphere post processor should be valid for vmware-iso builder")
	}

	// valid builder vmware-vmx
	valid = pp.IsValidWithBuilder("vmware-vmx")
	if !valid {
		t.Fatalf("vsphere post processor should be valid for vmware-vmx builder")
	}

	// invalid builder test
	valid = pp.IsValidWithBuilder("test")
	if valid {
		t.Fatalf("vsphere post processor should be valid for test builder")
	}

	// another post processor should be valid for any builder
	pp = &PostProcessor{
		Name: "test",
	}
	valid = pp.IsValidWithBuilder("test")
	if !valid {
		t.Fatalf("test post processor should be valid for test builder")
	}
}
