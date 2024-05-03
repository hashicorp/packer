// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput
package sleeper

import (
	"log"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	Duration string `mapstructure:"duration" required:"true"`
}

type Datasource struct {
	config       Config
	durationTime time.Duration
}

type DatasourceOutput struct {
	Status bool `mapstructure:"status"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	dt, err := time.ParseDuration(d.config.Duration)
	if err != nil {
		return err
	}

	d.durationTime = dt

	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	log.Printf("[sleeper] Sleeping for %s", d.config.Duration)
	time.Sleep(d.durationTime)
	log.Printf("[sleeper] Done sleeping!")

	output := DatasourceOutput{
		Status: true,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
