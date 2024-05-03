// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"

	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeTemplateFileFunc constructs a function that takes a file path and
// an arbitrary object of named values and attempts to render the referenced
// file as a template using HCL template syntax.
//
// The template itself may recursively call other functions so a callback
// must be provided to get access to those functions. The template cannot,
// however, access any variables defined in the scope: it is restricted only to
// those variables provided in the second function argument.
//
// As a special exception, a referenced template file may not recursively call
// the templatefile function, since that would risk the same file being
// included into itself indefinitely.
func MakeTemplateFileFunc(baseDir string, funcsCb func() map[string]function.Function) function.Function {

	params := []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
		{
			Name: "vars",
			Type: cty.DynamicPseudoType,
		},
	}

	loadTmpl := func(fn string) (hcl.Expression, error) {
		// We re-use File here to ensure the same filename interpretation
		// as it does, along with its other safety checks.
		tmplVal, err := filesystem.File(baseDir, cty.StringVal(fn))
		if err != nil {
			return nil, err
		}

		expr, diags := hclsyntax.ParseTemplate([]byte(tmplVal.AsString()), fn, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, diags
		}

		return expr, nil
	}

	renderTmpl := func(expr hcl.Expression, varsVal cty.Value) (cty.Value, error) {
		if varsTy := varsVal.Type(); !(varsTy.IsMapType() || varsTy.IsObjectType()) {
			return cty.DynamicVal, function.NewArgErrorf(1, "invalid vars value: must be a map") // or an object, but we don't strongly distinguish these most of the time
		}

		ctx := &hcl.EvalContext{
			Variables: varsVal.AsValueMap(),
		}

		// We require all of the variables to be valid HCL identifiers, because
		// otherwise there would be no way to refer to them in the template
		// anyway. Rejecting this here gives better feedback to the user
		// than a syntax error somewhere in the template itself.
		for n := range ctx.Variables {
			if !hclsyntax.ValidIdentifier(n) {
				// This error message intentionally doesn't describe _all_ of
				// the different permutations that are technically valid as an
				// HCL identifier, but rather focuses on what we might
				// consider to be an "idiomatic" variable name.
				return cty.DynamicVal, function.NewArgErrorf(1, "invalid template variable name %q: must start with a letter, followed by zero or more letters, digits, and underscores", n)
			}
		}

		// We'll pre-check references in the template here so we can give a
		// more specialized error message than HCL would by default, so it's
		// clearer that this problem is coming from a templatefile call.
		for _, traversal := range expr.Variables() {
			root := traversal.RootName()
			if _, ok := ctx.Variables[root]; !ok {
				return cty.DynamicVal, function.NewArgErrorf(1, "vars map does not contain key %q, referenced at %s", root, traversal[0].SourceRange())
			}
		}

		givenFuncs := funcsCb() // this callback indirection is to avoid chicken/egg problems
		funcs := make(map[string]function.Function, len(givenFuncs))
		for name, fn := range givenFuncs {
			if name == "templatefile" {
				// We stub this one out to prevent recursive calls.
				funcs[name] = function.New(&function.Spec{
					Params: params,
					Type: func(args []cty.Value) (cty.Type, error) {
						return cty.NilType, fmt.Errorf("cannot recursively call templatefile from inside templatefile call")
					},
				})
				continue
			}
			funcs[name] = fn
		}
		ctx.Functions = funcs

		val, diags := expr.Value(ctx)
		if diags.HasErrors() {
			return cty.DynamicVal, diags
		}
		return val, nil
	}

	return function.New(&function.Spec{
		Params: params,
		Type: func(args []cty.Value) (cty.Type, error) {
			if !(args[0].IsKnown() && args[1].IsKnown()) {
				return cty.DynamicPseudoType, nil
			}

			// We'll render our template now to see what result type it
			// produces. A template consisting only of a single interpolation
			// can potentially return any type.
			expr, err := loadTmpl(args[0].AsString())
			if err != nil {
				return cty.DynamicPseudoType, err
			}

			// This is safe even if args[1] contains unknowns because the HCL
			// template renderer itself knows how to short-circuit those.
			val, err := renderTmpl(expr, args[1])
			return val.Type(), err
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			expr, err := loadTmpl(args[0].AsString())
			if err != nil {
				return cty.DynamicVal, err
			}
			return renderTmpl(expr, args[1])
		},
	})

}
