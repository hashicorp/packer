//go:generate mapstructure-to-hcl2 -type MockDatasource,MockDatasourceResponse
package packer

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	configHelper "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type MockDatasource struct {
	Foo string

	OutputSpecCalled bool          `mapstructure-to-hcl2:",skip"`
	ConfigureCalled  bool          `mapstructure-to-hcl2:",skip"`
	ConfigureConfigs []interface{} `mapstructure-to-hcl2:",skip"`
	ExecuteCalled    bool          `mapstructure-to-hcl2:",skip"`
}

type MockDatasourceResponse struct {
	Foo string
}

func (d *MockDatasource) ConfigSpec() hcldec.ObjectSpec {
	return d.FlatMapstructure().HCL2Spec()
}

func (d *MockDatasource) OutputSpec() hcldec.ObjectSpec {
	d.OutputSpecCalled = true
	return (&MockDatasourceResponse{}).FlatMapstructure().HCL2Spec()
}

func (d *MockDatasource) Configure(configs ...interface{}) error {
	configHelper.Decode(d, nil, configs...)
	d.ConfigureCalled = true
	d.ConfigureConfigs = configs
	return nil
}

func (d *MockDatasource) Execute() (cty.Value, error) {
	d.ExecuteCalled = true
	if d.Foo == "" {
		d.Foo = "bar"
	}
	return cty.ObjectVal(map[string]cty.Value{
		"foo": cty.StringVal(d.Foo),
	}), nil
}
