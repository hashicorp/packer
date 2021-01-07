package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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

type DatasourceRef struct {
	Type string
	Name string
}

// the 'addition' field makes of ref a different entry in the data sources map, so
// Ref is here to make sure only one is returned.
func (r *DatasourceRef) Ref() DatasourceRef {
	return DatasourceRef{
		Type: r.Type,
		Name: r.Name,
	}
}

func (cfg *PackerConfig) startDatasource(dataSourceStore packer.DatasourceStore, ref DatasourceRef) (packersdk.Datasource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	block := cfg.Datasources[ref].block

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
	decoded, moreDiags := decodeHCL2Spec(body, cfg.EvalContext(nil), datasource)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}

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

	if p.DatasourceSchemas == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + r.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("packer does not currently know any data source."),
			Severity: hcl.DiagError,
		})
		return r, diags
	}

	if !p.DatasourceSchemas.Has(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + r.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known data sources: %v", p.DatasourceSchemas.List()),
			Severity: hcl.DiagError,
		})
		return r, diags
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
