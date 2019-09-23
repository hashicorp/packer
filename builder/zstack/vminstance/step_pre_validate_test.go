package vminstance

import (
	"testing"
)

func TestValidate_VersionCompare(t *testing.T) {
	cases := []struct {
		v1   string
		v2   string
		less bool
	}{
		{
			"3.3.0",
			"3.4.0",
			true,
		},
		{
			"3.3",
			"3.4.0",
			true,
		},
		{
			"2.6.0",
			"3.4.0",
			true,
		},
		{
			"2",
			"3.4.0",
			true,
		},
		{
			"3.4.0",
			"3.4.1",
			true,
		},
		{
			"2.5.0",
			"3",
			true,
		},
		{
			"3.5.0",
			"3.4.0",
			false,
		},
		{
			"3.4.1",
			"3.4.0",
			false,
		},
		{
			"3.4.0",
			"3.4.0",
			false,
		},
		{
			"3.4",
			"3.4.0",
			false,
		},
		{
			"3.5",
			"3.4.0",
			false,
		},
		{
			"4",
			"3.4.0",
			false,
		},
		{
			"3.5.4",
			"3",
			false,
		},
	}

	for _, c := range cases {
		r := versionLessThan(c.v1, c.v2)
		if r != c.less {
			t.Fatalf("compare version(%s, %s) failed!", c.v1, c.v2)
		}
	}
}
