package schedulers

import (
	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

func (s HCLSequentialScheduler) localVariablesEvaluationDone() bool {
	for _, loc := range s.config.LocalBlocks {
		if !loc.Evaluated() {
			return false
		}
	}

	return true
}

func (s *HCLSequentialScheduler) EvaluateVariables() hcl.Diagnostics {
	diags := s.config.InputVariables.ValidateValues()
	diags = diags.Extend(s.config.CheckForDuplicateLocalDefinition())

	diags = diags.Extend(s.evaluateVariables())

	filterVarsFromLogs(s.config.InputVariables)
	filterVarsFromLogs(s.config.LocalVariables)

	return diags
}

func (s *HCLSequentialScheduler) evaluateVariables() hcl.Diagnostics {
	if len(s.config.LocalBlocks) == 0 {
		return nil
	}

	// If we're done evaluating variables, we can leave immediately
	if s.localVariablesEvaluationDone() {
		return nil
	}

	var diags hcl.Diagnostics

	found := false
	for _, loc := range s.config.LocalBlocks {
		if loc.Evaluated() {
			continue
		}

		evalDiags := loc.Evaluate(s.config)
		// If there's a not ready for eval error, we continue iterating
		// on the other variable blocks, until we reach a point where
		// we can evaluate it.
		if evalDiags.HasErrors() &&
			evalDiags[0].Summary == hcl2template.VarNotReadyForEval {
			continue
		}

		found = true
	}

	if !found {
		return append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to evaluate local variables",
			Detail:   "Packer couldn't evaluate any more variables, and some are pending. This likely means your configuration has a dependency cycle in the local variables, that needs to be corrected before building the template.",
		})
	}

	return diags.Extend(s.evaluateVariables())
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
