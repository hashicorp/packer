package main

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/packer/command"
	"github.com/mitchellh/cli"
)

func TestExcludeHelpFunc(t *testing.T) {
	commands := map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
				Meta: command.Meta{},
			}, nil
		},

		"fix": func() (cli.Command, error) {
			return &command.FixCommand{
				Meta: command.Meta{},
			}, nil
		},
	}

	helpFunc := excludeHelpFunc(commands, []string{"fix"})
	helpText := helpFunc(commands)

	if strings.Contains(helpText, "fix") {
		t.Fatalf("Found fix in help text even though we excluded it: \n\n%s\n\n", helpText)
	}
}

func TestExtractChdir(t *testing.T) {
	cases := []struct {
		desc, overrideWd string
		args, expected   []string
		err              error
	}{
		{
			desc:       "TestHappyPath",
			args:       []string{"-chdir=example", "foo", "bar"},
			expected:   []string{"foo", "bar"},
			overrideWd: "example",
			err:        nil,
		},
		{
			desc:       "TestEmptyArgs",
			args:       []string{},
			expected:   []string{},
			overrideWd: "",
			err:        nil,
		},
		{
			desc:       "TestNoChdirArg",
			args:       []string{"foo", "bar"},
			expected:   []string{"foo", "bar"},
			overrideWd: "",
			err:        nil,
		},
		{
			desc:       "TestChdirNotFirst",
			args:       []string{"foo", "-chdir=example"},
			expected:   []string{"foo", "-chdir=example"},
			overrideWd: "",
			err:        nil,
		},
	}

	for _, tc := range cases {
		overrideWd, args, err := extractChdirOption(tc.args)
		if overrideWd != tc.overrideWd {
			t.Fatalf("%s: bad overrideWd,  expected: %s got: %s for args: %#v",
				tc.desc, tc.overrideWd, overrideWd, tc.args)
		}

		if !reflect.DeepEqual(args, tc.expected) {
			t.Fatalf("%s: bad result args, expected: %#v, got: %#v for args: %#v", tc.desc, tc.expected, args, tc.args)
		}

		if err != tc.err {
			t.Fatalf("%s: bad err, expected: %s, got: %s for args: %#v", tc.desc, tc.err, err, tc.args)
		}
	}
}

func TestExtractMachineReadable(t *testing.T) {
	var args, expected, result []string
	var mr bool

	// Not
	args = []string{"foo", "bar", "baz"}
	result, mr = extractMachineReadable(args)
	expected = []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}

	if mr {
		t.Fatal("should not be mr")
	}

	// Yes
	args = []string{"foo", "-machine-readable", "baz"}
	result, mr = extractMachineReadable(args)
	expected = []string{"foo", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}

	if !mr {
		t.Fatal("should be mr")
	}
}

func TestRandom(t *testing.T) {
	if rand.Intn(9999999) == 8498210 {
		t.Fatal("math.rand is not seeded properly")
	}
}
