package schedulers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template"
)

// datasourcesDone checks whether all the datasources have been executed or not
func (s *HCLSequentialScheduler) datasourcesDone() bool {
	for _, ds := range s.config.Datasources {
		if !ds.Executed() {
			return false
		}
	}

	return true
}

func (s *HCLSequentialScheduler) ExecuteDataSources(skipDatasourcesExecution bool) hcl.Diagnostics {
	// If we are done with datasources execution, we leave immediately
	if s.datasourcesDone() {
		return nil
	}

	var outDiags hcl.Diagnostics

	foundSomething := false
	for _, ds := range s.config.Datasources {
		if ds.Executed() {
			continue
		}

		diags := ds.Execute(s.config, skipDatasourcesExecution)
		if diags.HasErrors() && diags[0].Summary == hcl2template.NotReadyDataSourceError {
			// If we have a not ready error in the
			// datasource list, we should attempt to run the
			// rest, and eventually settle if we cannot move
			// any further
			continue
		}

		foundSomething = true

		outDiags = append(outDiags, diags...)
	}

	if !foundSomething {
		return append(outDiags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "No datasource could be executed",
			Detail: `While trying to recurisvely evaluating datasources, we could not find a next datasource to execute.
This is likely due to a cyclic dependency in your datasources`,
		})
	}

	// If we couldn't execute a datasource for whatever reason, we leave
	if outDiags.HasErrors() {
		return outDiags
	}

	// If we still found something to execute, we recursively execute the
	// remainder of the datasources
	return outDiags.Extend(s.ExecuteDataSources(skipDatasourcesExecution))
}
