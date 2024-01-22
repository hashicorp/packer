// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package file

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The contents of the file to create
	//
	// This is useful especially for files that involve templating so that
	// Packer can dynamically create files and expose them for later importing
	// as attributes in another entity.
	//
	// If no contents are specified, the resulting file will be empty.
	Contents string `mapstructure:"contents" required:"false"`
	// The file to write the contents to.
	Destination string `mapstructure:"destination" required:"true"`
	// Erase the destination if it exists.
	//
	// Default: `false`
	Force bool `mapstructure:"force" required:"false"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// The path of the file created
	Path string `mapstructure:"path"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	if d.config.Destination == "" {
		return fmt.Errorf("The `destination` must be specified.")
	}

	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	nulVal := cty.NullVal(cty.EmptyObject)

	_, err := os.Stat(d.config.Destination)
	if err == nil {
		if !d.config.Force {
			return nulVal, fmt.Errorf("destination file %q already exists", d.config.Destination)
		}
	}

	dest, err := os.OpenFile(d.config.Destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nulVal, fmt.Errorf("failed to create destination %q: %s", d.config.Destination, err)
	}

	defer dest.Close()

	written, err := dest.Write([]byte(d.config.Contents))
	if err != nil {
		defer os.Remove(d.config.Destination)
		return nulVal, fmt.Errorf("failed to write contents to %q: %s", d.config.Destination, err)
	}

	if written != len(d.config.Contents) {
		defer os.Remove(d.config.Destination)
		return nulVal, fmt.Errorf(
			"failed to write contents to %q: expected to write %d bytes, but wrote %d instead",
			d.config.Destination,
			len(d.config.Contents),
			written)
	}

	output := DatasourceOutput{
		Path: d.config.Destination,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
