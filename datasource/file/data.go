// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package file

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	// The file or directory to write the contents to.
	Destination string `mapstructure:"destination" required:"false"`
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

	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	nulVal := cty.NullVal(cty.EmptyObject)

	dest, err := d.createTempOutputFile()
	if err != nil {
		return nulVal, fmt.Errorf("failed to create output file: %s", err)
	}
	defer dest.Close()

	log.Printf("[INFO] data/file - Writing to %q", dest.Name())

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
		Path: dest.Name(),
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func (d *Datasource) createTempOutputFile() (*os.File, error) {
	// If we did not get a destination, we'll create a temp file in the
	// system's temporary directory
	if d.config.Destination == "" {
		return os.CreateTemp("", "")
	}

	// First try to stat the destination, to determine if it already exists and its type
	st, statErr := os.Stat(d.config.Destination)
	if statErr == nil {
		if st.IsDir() {
			return os.CreateTemp(d.config.Destination, "")
		}

		return os.OpenFile(d.config.Destination, os.O_TRUNC|os.O_RDWR, 0644)
	}

	outDir := filepath.Dir(d.config.Destination)

	// In case the destination does not exist, we'll get the dirpath,
	// and create it if it doesn't already exist
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination directory %q: %s", outDir, err)
	}

	// Check if the destination is a directory after the previous step.
	//
	// This happens if the path specified ends with a `/`, in which case the
	// destination is a directory, and we must create a temporary file in
	// this destination directory.
	destStat, statErr := os.Stat(d.config.Destination)
	if statErr == nil && destStat.IsDir() {
		return os.CreateTemp(d.config.Destination, "")
	}

	return os.Create(d.config.Destination)
}
