package packer

import "github.com/hashicorp/hcl/v2/hcldec"

// a struct (or type) implementing HCL2Speccer is a type that can tell it's own
// hcl2 conf/layout.
type HCL2Speccer interface {
	// ConfigSpec should return the hcl object spec used to configure the
	// builder. It will be used to tell the HCL parsing library how to
	// validate/configure a configuration.
	ConfigSpec() hcldec.ObjectSpec
}
