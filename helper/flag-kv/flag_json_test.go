package kvflag

import (
	"flag"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFlagJSON_impl(t *testing.T) {
	var _ flag.Value = new(FlagJSON)
}

func TestFlagJSON(t *testing.T) {
	cases := []struct {
		Input   string
		Initial map[string]string
		Output  map[string]string
		Error   bool
	}{
		{
			"basic.json",
			nil,
			map[string]string{"key": "value"},
			false,
		},

		{
			"basic.json",
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "bar", "key": "value"},
			false,
		},

		{
			"basic.json",
			map[string]string{"key": "bar"},
			map[string]string{"key": "value"},
			false,
		},
	}

	for _, tc := range cases {
		f := new(FlagJSON)
		if tc.Initial != nil {
			f = (*FlagJSON)(&tc.Initial)
		}

		err := f.Set(filepath.Join("./test-fixtures", tc.Input))
		if (err != nil) != tc.Error {
			t.Fatalf("bad error. Input: %#v\n\n%s", tc.Input, err)
		}

		actual := map[string]string(*f)
		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("bad: %#v", actual)
		}
	}
}
