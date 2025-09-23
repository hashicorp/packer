package function

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

func TestFilebase64(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectedOutput string
		expectError    bool
	}{
		{
			"file exists, return base64'd contents, no error",
			"./testdata/list.tmpl",
			"JXsgZm9yIHggaW4gbGlzdCB+fQotICR7eH0KJXsgZW5kZm9yIH59Cg==",
			false,
		},
		{
			"file doesn't exist, return nilval and an error",
			"./testdata/no_file",
			"",
			true,
		},
		{
			"directory passed as arg, should error",
			"./testdata",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Filebase64.Call([]cty.Value{
				cty.StringVal(tt.file),
			})

			if tt.expectError && err == nil {
				t.Fatal("succeeded; want error")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if err != nil {
				return
			}

			retVal := res.AsString()
			diff := cmp.Diff(retVal, tt.expectedOutput)
			if diff != "" {
				t.Errorf("expected output and returned are different: %s", diff)
			}
		})
	}
}
