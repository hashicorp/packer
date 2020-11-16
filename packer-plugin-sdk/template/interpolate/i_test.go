package interpolate

import (
	"testing"
)

func TestIRender(t *testing.T) {
	cases := map[string]struct {
		Ctx    *Context
		Value  string
		Result string
	}{
		"basic": {
			nil,
			"foo",
			"foo",
		},
	}

	for k, tc := range cases {
		i := &I{Value: tc.Value}
		result, err := i.Render(tc.Ctx)
		if err != nil {
			t.Fatalf("%s\n\ninput: %s\n\nerr: %s", k, tc.Value, err)
		}
		if result != tc.Result {
			t.Fatalf(
				"%s\n\ninput: %s\n\nexpected: %s\n\ngot: %s",
				k, tc.Value, tc.Result, result)
		}
	}
}
