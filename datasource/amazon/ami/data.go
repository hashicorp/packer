//go:generate mapstructure-to-hcl2 -type DatasourceOutput
package ami

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	ID           string
	Name         string
	CreationDate string
	Owner        string
	OwnerName    string
	Tags         map[string]string
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(...interface{}) error {
	return nil
}

func (d *Datasource) Execute() (cty.Value, error) {
	return cty.ObjectVal(map[string]cty.Value{
		"id": cty.StringVal("ami-0568456c"),
	}), nil
}
