// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// DatasourceBlock references an HCL 'data' block.
type DatasourceBlock struct {
	Type string
	Name string

	Value cty.Value
	Block *hcl.Block
}

type DatasourceRef struct {
	Type string
	Name string
}

type Datasources map[DatasourceRef]DatasourceBlock

func (data *DatasourceBlock) Ref() DatasourceRef {
	return DatasourceRef{
		Type: data.Type,
		Name: data.Name,
	}
}

func (ds *Datasources) Values() (map[string]cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := map[string]cty.Value{}
	valuesMap := map[string]map[string]cty.Value{}

	for ref, datasource := range *ds {
		if datasource.Value == (cty.Value{}) {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  fmt.Sprintf("empty value"),
				Subject:  &datasource.Block.DefRange,
				Severity: hcl.DiagError,
			})
			continue
		}
		inner := valuesMap[ref.Type]
		if inner == nil {
			inner = map[string]cty.Value{}
		}
		inner[ref.Name] = datasource.Value
		res[ref.Type] = cty.MapVal(inner)

		// Keeps values of different datasources from same type
		valuesMap[ref.Type] = inner
	}

	return res, diags
}

func (cfg *PackerConfig) StartDatasource(dataSourceStore packer.DatasourceStore, ref DatasourceRef, secondaryEvaluation bool) (packersdk.Datasource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	block := cfg.Datasources[ref].Block

	if dataSourceStore == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + DataSourceLabel + " type " + ref.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("packer does not currently know any data source."),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	if !dataSourceStore.Has(ref.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + DataSourceLabel + " type " + ref.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known data sources: %v", dataSourceStore.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	datasource, err := dataSourceStore.Start(ref.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}
	if datasource == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  fmt.Sprintf("failed to start datasource plugin %q.%q", ref.Type, ref.Name),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}

	var decoded cty.Value
	var moreDiags hcl.Diagnostics
	body := block.Body
	decoded, moreDiags = DecodeHCL2Spec(body, cfg.EvalContext(nil), datasource)

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

func (cfg *PackerConfig) ExecuteDatasource(ref DatasourceRef, skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	ds := cfg.Datasources[ref]

	datasource, startDiags := cfg.StartDatasource(cfg.Parser.PluginConfig.DataSources, ref, false)
	diags = append(diags, startDiags...)
	if diags.HasErrors() {
		return diags
	}

	// If we run validate, we want to boot the datasource binary, but we don't
	// want to actually execute it, so we replace the value with a placeholder
	// and immediately return.
	if skipExecution {
		placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
		ds.Value = placeholderValue
		cfg.Datasources[ref] = ds
		return diags
	}

	dsOpts, _ := DecodeHCL2Spec(ds.Block.Body, cfg.EvalContext(nil), datasource)
	sp := packer.CheckpointReporter.AddSpan(ref.Type, "datasource", dsOpts)
	realValue, err := datasource.Execute()
	sp.End(err)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &cfg.Datasources[ref].Block.DefRange,
			Severity: hcl.DiagError,
			Context:  &ds.Block.DefRange,
		})
		return diags
	}

	ds.Value = realValue
	cfg.Datasources[ref] = ds

	return diags
}

func (p *Parser) decodeDataBlock(block *hcl.Block) (*DatasourceBlock, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	r := &DatasourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		Block: block,
	}

	if !hclsyntax.ValidIdentifier(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data source name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}
	if !hclsyntax.ValidIdentifier(r.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data resource name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[1],
		})
	}

	return r, diags
}
