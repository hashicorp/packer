package packer

import "github.com/zclconf/go-cty/cty"

type DataSource interface {
	// HCL2Speccer is a type that can tell it's own hcl2 conf/layout.
	HCL2Speccer

	// Configure takes values from HCL2 and applies them to the struct
	Configure(...interface{}) error

	// Execute the func call and return the values
	Execute() (cty.Value, error)
}
