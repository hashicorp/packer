package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/zclconf/go-cty/cty"
)

// MetaBlock defines funcs that are common to all HCL2 blocks.
type MetaBlock interface {
	// References returns the *addrs.Reference of everything being referenced in
	// this block.
	References() ([]*addrs.Reference, hcl.Diagnostics)

	// Evaluate runs HCL all expression and puts them into the underlying data
	// model.
	//
	// For a block that is not a Runner — like a variable block — this will make
	// Value known.
	Evaluate(ctx *hcl.EvalContext) hcl.Diagnostics

	// Value of the block. For non-Runner types it is known from the top,
	// otherwise it is probably unknown or not fully known before Run is called.
	Value() cty.Value

	// Expected type of Value, when unknown: cty.DynamicPseudoType is returned.
	Type() cty.Type
}

type Runner interface {
	// Run will be available on blocks that can execute, like the data source
	// block or a build block.
	//
	// The run should result in changing the respone of Value to something
	// known.
	Run() hcl.Diagnostics
}

var (
	_ MetaBlock = &Variable{}
	_ MetaBlock = &DatasourceBlock{}
	_ Runner    = &DatasourceBlock{}
	_ MetaBlock = &BuildBlock{}
	_ Runner    = &BuildBlock{}
)
