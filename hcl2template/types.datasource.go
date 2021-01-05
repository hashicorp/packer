package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

// DataBlock references an HCL 'data' block.
type DataSource struct {
	Type string
	Name string

	value cty.Value
	block *hcl.Block
}

type DataSources map[DataSourceRef]DataSource

func (data *DataSource) Ref() DataSourceRef {
	return DataSourceRef{
		Type: data.Type,
		Name: data.Name,
	}
}

func (ds *DataSources) Values() (map[string]cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := map[string]cty.Value{}

	for ref, datasource := range *ds {
		if datasource.value == (cty.Value{}) {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  fmt.Sprintf("empty value"),
				Subject:  &datasource.block.DefRange,
				Severity: hcl.DiagError,
			})
			continue
		}

		inner := map[string]cty.Value{}
		inner[ref.Name] = datasource.value
		res[ref.Type] = cty.MapVal(inner)
	}

	return res, diags
}

type DataSourceRef struct {
	Type string
	Name string
}

// the 'addition' field makes of ref a different entry in the data sources map, so
// Ref is here to make sure only one is returned.
func (r *DataSourceRef) Ref() DataSourceRef {
	return DataSourceRef{
		Type: r.Type,
		Name: r.Name,
	}
}

func (cfg *PackerConfig) startDatasource(dataSourceStore packer.DataSourceStore, ref DataSourceRef) (packersdk.DataSource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	block := cfg.DataSources[ref].block

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

func (p *Parser) decodeDataBlock(block *hcl.Block) (*DataSource, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	r := &DataSource{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}

	if p.DataSourceSchemas == nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + r.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("packer does not currently know any data source."),
			Severity: hcl.DiagError,
		})
		return r, diags
	}

	if !p.DataSourceSchemas.Has(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + dataSourceLabel + " type " + r.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known data sources: %v", p.DataSourceSchemas.List()),
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

func getSpecValue(spec hcldec.Spec) cty.Value {
	switch spec := spec.(type) {
	case *hcldec.LiteralSpec:
		return spec.Value
	}
	return cty.UnknownVal(hcldec.ImpliedType(spec))
}
