package atlas

import (
	"path/filepath"
	"testing"
)

func TestLongestCommonPrefix(t *testing.T) {
	sep := string(filepath.Separator)
	cases := []struct {
		Input  []string
		Output string
	}{
		{
			[]string{"foo", "bar"},
			"",
		},
		{
			[]string{"foo", "foobar"},
			"",
		},
		{
			[]string{"foo" + sep, "foo" + sep + "bar"},
			"foo" + sep,
		},
		{
			[]string{sep + "foo" + sep, sep + "bar"},
			sep,
		},
	}

	for _, tc := range cases {
		actual := longestCommonPrefix(tc.Input)
		if actual != tc.Output {
			t.Fatalf("bad: %#v\n\n%#v", actual, tc.Input)
		}
	}
}
