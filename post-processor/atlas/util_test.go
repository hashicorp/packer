package atlas

import (
	"testing"
)

func TestLongestCommonPrefix(t *testing.T) {
	cases := []struct {
		Input []string
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
			[]string{"foo/", "foo/bar"},
			"foo/",
		},
		{
			[]string{"/foo/", "/bar"},
			"/",
		},
	}

	for _, tc := range cases {
		actual := longestCommonPrefix(tc.Input)
		if actual != tc.Output {
			t.Fatalf("bad: %#v\n\n%#v", actual, tc.Input)
		}
	}
}
