package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// DataBlock references an HCL 'data' block.
type Datasource struct {
	Type string
	Name string

	value cty.Value
	block *hcl.Block
}

type DatasourceRef struct {
	Type string
	Name string
}

type Datasources map[DatasourceRef]Datasource

func (data *Datasource) Ref() DatasourceRef {
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
		if datasource.value == (cty.Value{}) {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  fmt.Sprintf("empty value"),
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

func (cfg *PackerConfig) startDatasource(dataSourceStore packer.DatasourceStore, ref DatasourceRef) (packersdk.Datasource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	block := cfg.Datasources[ref].block

	if dataSourceStore == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + ref.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("packer does not currently know any data source."),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	if !dataSourceStore.Has(ref.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + ref.Type,
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
	body := block.Body
	decoded, moreDiags := decodeHCL2Spec(body, cfg.EvalContext(DatasourceContext, nil), datasource)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

	// In case of cty.Unknown values, this will write a equivalent placeholder of the same type
	// Unknown types are not recognized by the json marshal during the RPC call and we have to do this here
	// to avoid json parsing failures when running the validate command.
	// We don't do this before so we can validate if variable types matches correctly on decodeHCL2Spec.
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

func (p *Parser) decodeDataBlock(block *hcl.Block) (*Datasource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	r := &Datasource{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
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
