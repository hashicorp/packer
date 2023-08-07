package schedulers

import (
	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

func (s *HCLSequentialScheduler) EvaluateVariables() hcl.Diagnostics {
	locals := s.config.LocalBlocks

	var diags hcl.Diagnostics

	if len(locals) == 0 {
		return diags
	}

	if s.config.LocalVariables == nil {
		s.config.LocalVariables = hcl2template.Variables{}
	}

	for foundSomething := true; foundSomething; {
		foundSomething = false
		for i := 0; i < len(locals); {
			local := locals[i]
			moreDiags := s.EvaluateLocalVariable(local)
			if moreDiags.HasErrors() {
				i++
				continue
			}
			foundSomething = true
			locals = append(locals[:i], locals[i+1:]...)
		}
	}

	if len(locals) != 0 {
		// get errors from remaining variables
		return s.EvaluateAllLocalVariables(locals)
	}

	filterVarsFromLogs(s.config.InputVariables)
	filterVarsFromLogs(s.config.LocalVariables)

	return diags
}

func (s *HCLSequentialScheduler) EvaluateLocalVariable(local *hcl2template.LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	value, moreDiags := local.Expr.Value(s.config.EvalContext(nil))
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return diags
	}
	s.config.LocalVariables[local.Name] = &hcl2template.Variable{
		Name:      local.Name,
		Sensitive: local.Sensitive,
		Values: []hcl2template.VariableAssignment{{
			Value: value,
			Expr:  local.Expr,
			From:  "default",
		}},
		Type: value.Type(),
	}

	return diags
}

func (s *HCLSequentialScheduler) EvaluateAllLocalVariables(locals []*hcl2template.LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, local := range locals {
		diags = append(diags, s.EvaluateLocalVariable(local)...)
	}

	return diags
}

func filterVarsFromLogs(inputOrLocal hcl2template.Variables) {
	for _, variable := range inputOrLocal {
		if !variable.Sensitive {
			continue
		}
		value := variable.Value()
		_ = cty.Walk(value, func(_ cty.Path, nested cty.Value) (bool, error) {
			if nested.IsWhollyKnown() && !nested.IsNull() && nested.Type().Equals(cty.String) {
				packersdk.LogSecretFilter.Set(nested.AsString())
			}
			return true, nil
		})
	}
}
