// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestTemplateFile(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Vars cty.Value
		Want cty.Value
		Err  string
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.EmptyObjectVal,
			cty.StringVal("Hello World"),
			``,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.EmptyObjectVal,
			cty.NilVal,
			`contents of testdata/icon.png are not valid UTF-8; use the filebase64 function to obtain the Base64 encoded contents or the other file functions (e.g. filemd5, filesha256) to obtain file hashing results instead`,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.EmptyObjectVal,
			cty.NilVal,
			`no file exists at ` + filepath.Clean("testdata/missing"),
		},
		{
			cty.StringVal("testdata/hello.tmpl"),
			cty.MapVal(map[string]cty.Value{
				"name": cty.StringVal("Jodie"),
			}),
			cty.StringVal("Hello, Jodie!"),
			``,
		},
		{
			cty.StringVal("testdata/hello.tmpl"),
			cty.MapVal(map[string]cty.Value{
				"name!": cty.StringVal("Jodie"),
			}),
			cty.NilVal,
			`invalid template variable name "name!": must start with a letter, followed by zero or more letters, digits, and underscores`,
		},
		{
			cty.StringVal("testdata/hello.tmpl"),
			cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Jimbo"),
			}),
			cty.StringVal("Hello, Jimbo!"),
			``,
		},
		{
			cty.StringVal("testdata/hello.tmpl"),
			cty.EmptyObjectVal,
			cty.NilVal,
			`vars map does not contain key "name", referenced at testdata/hello.tmpl:1,10-14`,
		},
		{
			cty.StringVal("testdata/func.tmpl"),
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
			}),
			cty.StringVal("The items are a, b, c"),
			``,
		},
		{
			cty.StringVal("testdata/recursive.tmpl"),
			cty.MapValEmpty(cty.String),
			cty.NilVal,
			`testdata/recursive.tmpl:1,3-16: Error in function call; Call to function "templatefile" failed: cannot recursively call templatefile from inside templatefile call.`,
		},
		{
			cty.StringVal("testdata/list.tmpl"),
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
			}),
			cty.StringVal("- a\n- b\n- c\n"),
			``,
		},
		{
			cty.StringVal("testdata/list.tmpl"),
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.True,
			}),
			cty.NilVal,
			`testdata/list.tmpl:1,13-17: Iteration over non-iterable value; A value of type bool cannot be used as the collection in a 'for' expression.`,
		},
		{
			cty.StringVal("testdata/bare.tmpl"),
			cty.ObjectVal(map[string]cty.Value{
				"val": cty.True,
			}),
			cty.True, // since this template contains only an interpolation, its true value shines through
			``,
		},
	}

	templateFileFn := MakeTemplateFileFunc(".", func() map[string]function.Function {
		return map[string]function.Function{
			"join":         stdlib.JoinFunc,
			"templatefile": filesystem.MakeFileFunc(".", false), // just a placeholder, since templatefile itself overrides this
		}
	})

	for _, test := range tests {
		t.Run(fmt.Sprintf("TemplateFile(%#v, %#v)", test.Path, test.Vars), func(t *testing.T) {
			got, err := templateFileFn.Call([]cty.Value{test.Path, test.Vars})

			if argErr, ok := err.(function.ArgError); ok {
				if argErr.Index < 0 || argErr.Index > 1 {
					t.Errorf("ArgError index %d is out of range for templatefile (must be 0 or 1)", argErr.Index)
				}
			}

			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
