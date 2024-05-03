// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NestedFirst,NestedSecond,DatasourceOutput
package dynamic

import (
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type NestedSecond struct {
	Name string `mapstructure:"name" required:"true"`
}

type NestedFirst struct {
	Name    string         `mapstructure:"name" required:"true"`
	Nesteds []NestedSecond `mapstructure:"extra" required:"false"`
}

type Config struct {
	Nesteds []NestedFirst `mapstructure:"extra" required:"false"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	Status string `mapstructure:"data"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	log.Printf("[DATASOURCE-DYNAMIC] Executing datasource dynamic")
	for _, nest := range d.config.Nesteds {
		log.Printf("[DATASOURCE-DYNAMIC] - First nest: %s", nest.Name)
		for _, nestedConfig := range nest.Nesteds {
			log.Printf("[DATASOURCE-DYNAMIC] - Second nest: %s.%s", nest.Name, nestedConfig.Name)
		}
	}

	output := DatasourceOutput{
		Status: "OK",
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
