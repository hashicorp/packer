// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/zclconf/go-cty/cty"
)

// Decodable structs are structs that can tell their hcl2 ObjectSpec; this
// config spec will be passed to hcldec.Decode and the result will be a
// cty.Value. This Value can then be applied on the said struct.
type Decodable interface {
	ConfigSpec() hcldec.ObjectSpec
}

// legacyTemplateExpression matches one expression of the legacy template
// engine, for example `{{timestamp}}` or `{{isotime "2006-01-02"}}`.
var legacyTemplateExpression = regexp.MustCompile(`\{\{[^{}]*\}\}`)

// legacyTimeFunctions are the functions of the legacy template engine whose
// result depends on the time at which they are evaluated.
var legacyTimeFunctions = []string{"timestamp", "isotime", "strftime"}

// decodeHCL2Spec decodes the HCL2 body into the config spec of the component
// and resolves the time dependent functions of the legacy template engine.
//
// HCL2 templates are evaluated by the Packer process, which then hands the
// resulting values to the plugins. Each plugin resolves legacy template
// expressions against the time its own process started, so the same
// expression can produce different values in a builder, a provisioner and a
// post-processor. Resolving the time dependent functions here, in the Packer
// process, keeps them identical for the whole build.
func decodeHCL2Spec(body hcl.Body, ectx *hcl.EvalContext, dec Decodable) (cty.Value, hcl.Diagnostics) {
	decoded, diags := hcldec.Decode(body, dec.ConfigSpec(), ectx)
	if diags.HasErrors() || decoded == cty.NilVal || decoded.IsNull() {
		return decoded, diags
	}

	decoded, err := cty.Transform(decoded, func(_ cty.Path, value cty.Value) (cty.Value, error) {
		unmarked, marks := value.Unmark()
		if unmarked.IsNull() || !unmarked.IsKnown() || !unmarked.Type().Equals(cty.String) {
			return value, nil
		}

		str := unmarked.AsString()
		rendered := renderLegacyTimeFunctions(str)
		if rendered == str {
			return value, nil
		}
		return cty.StringVal(rendered).WithMarks(marks), nil
	})
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to render time functions",
			Detail:   err.Error(),
		})
	}

	return decoded, diags
}

// renderLegacyTimeFunctions resolves the time dependent functions of the
// legacy template engine in str. Every other template expression is left
// untouched so that the plugin can resolve it at run time, when the build
// variables it refers to are available.
func renderLegacyTimeFunctions(str string) string {
	if !strings.Contains(str, "{{") {
		return str
	}

	return legacyTemplateExpression.ReplaceAllStringFunc(str, func(expression string) string {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expression, "{{"), "}}"))
		if !isLegacyTimeFunction(name) {
			return expression
		}

		rendered, err := interpolate.Render(expression, nil)
		if err != nil {
			// Leave expressions that the legacy engine cannot evaluate
			// untouched, as the plugin would report the error anyway.
			return expression
		}
		return rendered
	})
}

func isLegacyTimeFunction(name string) bool {
	for _, fn := range legacyTimeFunctions {
		if name == fn || strings.HasPrefix(name, fn+" ") || strings.HasPrefix(name, fn+"\t") {
			return true
		}
	}
	return false
}
