// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package null

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

type Datasource struct {
	config Config
}

// The Null data source is designed to demonstrate how data sources work, and
// to provide a test plugin. It does not do anything useful; you assign an
// input string and it gets returned as an output string.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// This variable will get stored as "output" in the output spec.
	Input string `mapstructure:"input" required:"true"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.Input == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The `input` must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	// Output will return the input variable, as output.
	Output string `mapstructure:"output"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	// Pass input variable through to output.
	output := DatasourceOutput{
		Output: d.config.Input,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
