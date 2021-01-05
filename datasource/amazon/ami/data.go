package ami

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type DataSource struct {
	config Config
}

func (d *DataSource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *DataSource) OutputSpec() hcldec.ObjectSpec {
	return hcldec.ObjectSpec{
		"amazon-ami": &hcldec.LiteralSpec{Value: cty.StringVal("placeholder")},
	}
}

func (d *DataSource) Configure(...interface{}) error {
	return nil
}

func (d *DataSource) Execute() (cty.Value, error) {
	return cty.StringVal("ami-0568456c"), nil
}
