package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// DataBlock references an HCL 'data' block.
type DataSource struct {
	Type string
	Name string

	block *hcl.Block
}

func (data *DataSource) Ref() DataSourceRef {
	return DataSourceRef{
		Type: data.Type,
		Name: data.Name,
	}
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
