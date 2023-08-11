package schedulers

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

func (s *HCLSequentialScheduler) EvaluateDataSources() hcl.Diagnostics {
	return s.evaluateDatasources()
}

func (s *HCLSequentialScheduler) evaluateDatasources() hcl.Diagnostics {
	var diags hcl.Diagnostics

	dependencies := map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef{}
	for ref, ds := range s.config.Datasources {
		if ds.Value != (cty.Value{}) {
			continue
		}
		// Pre-examine body of this data source to see if it uses another data
		// source in any of its input expressions. If so, skip evaluating it for
		// now, and add it to a list of datasources to evaluate again, later,
		// with the datasources in its context.
		// This is essentially creating a very primitive DAG just for data
		// source interdependencies.
		block := ds.Block
		body := block.Body
		attrs, _ := body.JustAttributes()

		skipFirstEval := false
		for _, attr := range attrs {
			vars := attr.Expr.Variables()
			for _, v := range vars {
				// check whether the variable is a data source
				if v.RootName() == "data" {
					// construct, backwards, the data source type and name we
					// need to evaluate before this one can be evaluated.
					dependsOn := hcl2template.DatasourceRef{
						Type: v[1].(hcl.TraverseAttr).Name,
						Name: v[2].(hcl.TraverseAttr).Name,
					}
					log.Printf("The data source %#v depends on datasource %#v", ref, dependsOn)
					if dependencies[ref] != nil {
						dependencies[ref] = append(dependencies[ref], dependsOn)
					} else {
						dependencies[ref] = []hcl2template.DatasourceRef{dependsOn}
					}
					skipFirstEval = true
				}
			}
		}

		// Now we have a list of data sources that depend on other data sources.
		// Don't evaluate these; only evaluate data sources that we didn't
		// mark  as having dependencies.
		if skipFirstEval {
			continue
		}

		diags = append(diags, s.config.ExecuteDatasource(ref, s.opts.SkipDatasourcesExecution)...)
	}

	// Now that most of our data sources have been started and executed, we can
	// try to execute the ones that depend on other data sources.
	for ref := range dependencies {
		_, moreDiags := s.recursivelyEvaluateDatasources(ref, dependencies, 0)
		// Deduplicate diagnostics to prevent recursion messes.
		cleanedDiags := map[string]*hcl.Diagnostic{}
		for _, diag := range moreDiags {
			cleanedDiags[diag.Summary] = diag
		}

		for _, diag := range cleanedDiags {
			diags = append(diags, diag)
		}
	}

	return diags
}

func (s *HCLSequentialScheduler) recursivelyEvaluateDatasources(
	ref hcl2template.DatasourceRef,
	dependencies map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef,
	depth int,
) (map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var moreDiags hcl.Diagnostics

	if depth > 10 {
		// Add a comment about recursion.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Max datasource recursion depth exceeded.",
			Detail: "An error occured while recursively evaluating data " +
				"sources. Either your data source depends on more than ten " +
				"other data sources, or your data sources have a cyclic " +
				"dependency. Please simplify your config to continue. ",
		})
		return dependencies, diags
	}

	// Make sure everything ref depends on has already been evaluated.
	for _, dep := range dependencies[ref] {
		if _, ok := dependencies[dep]; ok {
			depth += 1
			// If this dependency is not in the map, it means we've already
			// launched and executed this datasource. Otherwise, it means
			// we still need to run it. RECURSION TIME!!
			dependencies, moreDiags = s.recursivelyEvaluateDatasources(dep, dependencies, depth)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				return dependencies, diags
			}
		}
	}

	diags = append(diags, s.config.ExecuteDatasource(ref, s.opts.SkipDatasourcesExecution)...)

	// remove ref from the dependencies map.
	delete(dependencies, ref)
	return dependencies, diags
}
