// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// DatasourceBlock references an HCL 'data' block.
type DatasourceBlock struct {
	Type string
	Name string

	value cty.Value
	block *hcl.Block

	// dependencies is the list of datasources to execute before this one
	dependencies []DatasourceRef
}

type DatasourceRef struct {
	Type string
	Name string
}

type Datasources map[DatasourceRef]*DatasourceBlock

func (data *DatasourceBlock) Ref() DatasourceRef {
	return DatasourceRef{
		Type: data.Type,
		Name: data.Name,
	}
}

func (ds *DatasourceBlock) getDependencies() {
	var dependencies []DatasourceRef

	// Note: when looking at the expressions, we only need to care about
	// attributes, as HCL2 expressions are not allowed in a block's labels.
	vars := GetVarsByType(ds.block, "data")
	for _, v := range vars {
		// construct, backwards, the data source type and name we
		// need to evaluate before this one can be evaluated.
		dependencies = append(dependencies, DatasourceRef{
			Type: v[1].(hcl.TraverseAttr).Name,
			Name: v[2].(hcl.TraverseAttr).Name,
		})
	}

	ds.dependencies = dependencies
}

const notReadyDataSourceError = "Dependencies not ready"

// executed returns whether or not the datasource was executed
//
// Having a non-empty cty.Value object means this was filled-up after the
// datasource has been executed, so this is what we use for this test.
func (ds DatasourceBlock) executed() bool {
	return ds.value != cty.Value{}
}

// Execute starts the datasource and executes it immediately.
//
// If its dependencies are not ready for execution, this will return an error
// and will only execute when all its dependencies have executed.
func (ds *DatasourceBlock) Execute(cfg *PackerConfig, skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	ok := true
	for _, depRef := range ds.dependencies {
		dep := cfg.Datasources[depRef]
		if dep == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Nonexistent dependency referenced",
				Detail:   fmt.Sprintf("The referenced datasource %s.%s is not defined in the configuration, so this datasource won't be able to execute.", dep.Type, dep.Name),
				Subject:  &ds.block.DefRange,
			})
			ok = false
			continue
		}

		if !dep.executed() {
			ok = false
		}
	}

	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  notReadyDataSourceError,
			Detail:   "At least one dependency for the datasource is not executed already",
			Subject:  &ds.block.DefRange,
		})
		return diags
	}

	// If we've gotten here, then it means ref doesn't seem to have any further
	// dependencies we need to evaluate first. Evaluate it, with the cfg's full
	// data source context.
	datasource, diags := cfg.startDatasource(*ds)
	if diags.HasErrors() {
		return diags
	}

	if skipExecution {
		placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
		ds.value = placeholderValue
		return diags
	}

	opts, _ := decodeHCL2Spec(ds.block.Body, cfg.EvalContext(DatasourceContext, nil), datasource)
	sp := packer.CheckpointReporter.AddSpan(ds.Type, "datasource", opts)
	realValue, err := datasource.Execute()
	sp.End(err)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &ds.block.DefRange,
			Severity: hcl.DiagError,
		})
		return diags
	}

	ds.value = realValue
	return diags
}

func (ds *Datasources) Values() (map[string]cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := map[string]cty.Value{}
	valuesMap := map[string]map[string]cty.Value{}

	for ref, datasource := range *ds {
		if datasource.value == (cty.Value{}) {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  "empty value",
				Subject:  &datasource.block.DefRange,
				Severity: hcl.DiagError,
			})
			continue
		}
		inner := valuesMap[ref.Type]
		if inner == nil {
			inner = map[string]cty.Value{}
		}
		inner[ref.Name] = datasource.value
		res[ref.Type] = cty.MapVal(inner)

		// Keeps values of different datasources from same type
		valuesMap[ref.Type] = inner
	}

	return res, diags
}

func (cfg *PackerConfig) startDatasource(ds DatasourceBlock) (packersdk.Datasource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	block := ds.block

	dataSourceStore := cfg.parser.PluginConfig.DataSources

	if dataSourceStore == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + ds.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   "packer does not currently know any data source.",
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	if !dataSourceStore.Has(ds.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + ds.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known data sources: %v", dataSourceStore.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	datasource, err := dataSourceStore.Start(ds.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}
	if datasource == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  fmt.Sprintf("failed to start datasource plugin %q.%q", ds.Type, ds.Name),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}

	var decoded cty.Value
	var moreDiags hcl.Diagnostics
	body := block.Body
	decoded, moreDiags = decodeHCL2Spec(body, cfg.EvalContext(DatasourceContext, nil), datasource)

	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	// In case of cty.Unknown values, this will write a equivalent placeholder
	// of the same type. Unknown types are not recognized by the json marshal
	// during the RPC call and we have to do this here to avoid json parsing
	// failures when running the validate command. We don't do this before so
	// we can validate if variable type matches correctly on decodeHCL2Spec.
	decoded = hcl2shim.WriteUnknownPlaceholderValues(decoded)
	if err := datasource.Configure(decoded); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}
	return datasource, diags
}

// datasourcesDone checks whether all the datasources have been executed or not
func (cfg *PackerConfig) datasourcesDone() bool {
	for _, ds := range cfg.Datasources {
		if !ds.executed() {
			return false
		}
	}

	return true
}

func (cfg *PackerConfig) executeDatasources(skipExecution bool) hcl.Diagnostics {
	// If we are done with datasources execution, we leave immediately
	if cfg.datasourcesDone() {
		return nil
	}

	var outDiags hcl.Diagnostics

	foundSomething := false
outerDSEval:
	for _, ds := range cfg.Datasources {
		if ds.executed() {
			continue
		}

		diags := ds.Execute(cfg, skipExecution)
		for _, diag := range diags {
			if diag.Summary == notReadyDataSourceError {
				// If we have a not ready error in the
				// datasource list, we should attempt to run the
				// rest, and eventually settle if we cannot move
				// any further
				continue outerDSEval
			}
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
	return outDiags.Extend(cfg.executeDatasources(skipExecution))
}
