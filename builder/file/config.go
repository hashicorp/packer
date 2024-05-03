// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package file

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

var ErrTargetRequired = fmt.Errorf("target required")
var ErrContentSourceConflict = fmt.Errorf("Cannot specify source file AND content")

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Source  string `mapstructure:"source"`
	Target  string `mapstructure:"target"`
	Content string `mapstructure:"content"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	warnings := []string{}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return warnings, err
	}

	var errs *packersdk.MultiError

	if c.Target == "" {
		errs = packersdk.MultiErrorAppend(errs, ErrTargetRequired)
	}

	if c.Content == "" && c.Source == "" {
		warnings = append(warnings, "Both source file and contents are blank; target will have no content")
	}

	if c.Content != "" && c.Source != "" {
		errs = packersdk.MultiErrorAppend(errs, ErrContentSourceConflict)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
