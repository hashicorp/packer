//go:generate mapstructure-to-hcl2 -type DataSourceOutput
package ami

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type DataSource struct {
	config Config
}

type DataSourceOutput struct {
	ID           string
	Name         string
	CreationDate string
	Owner        string
	OwnerName    string
	Tags         map[string]string
}

func (d *DataSource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *DataSource) OutputSpec() hcldec.ObjectSpec {
	return (&DataSourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *DataSource) Configure(...interface{}) error {
	return nil
}

func (d *DataSource) Execute() (cty.Value, error) {
	return cty.ObjectVal(map[string]cty.Value{
		"id": cty.StringVal("ami-0568456c"),
	}), nil
}
