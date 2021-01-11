package packer

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// Datasources make data available for use in any source block of a Packer configuration.
type Datasource interface {
	// HCL2Speccer is a type that can tell it's own hcl2 conf/layout.
	HCL2Speccer

	// Configure takes values from HCL2 and applies them to the struct
	Configure(...interface{}) error

	// OutputSpec is the HCL2 layout of the variable output, it will allow
	// Packer to validate whether someone is using the output of the data
	// source correctly without having to execute the data source call.
	OutputSpec() hcldec.ObjectSpec

	// Execute the func call and return the values
	Execute() (cty.Value, error)
}
