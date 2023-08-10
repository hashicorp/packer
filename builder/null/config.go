// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package null

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CommConfig communicator.Config `mapstructure:",squash"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	err := config.Decode(c, &config.DecodeOpts{
		PluginType:        BuilderId,
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packersdk.MultiError
	if es := c.CommConfig.Prepare(nil); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	if c.CommConfig.Type != "none" {
		if c.CommConfig.Host() == "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("a Host must be specified, please reference your communicator documentation"))
		}

		if c.CommConfig.User() == "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("a Username must be specified, please reference your communicator documentation"))
		}

		if !c.CommConfig.SSHAgentAuth && c.CommConfig.Password() == "" && c.CommConfig.SSHPrivateKeyFile == "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("one authentication method must be specified, please reference your communicator documentation"))
		}

		if (c.CommConfig.SSHAgentAuth &&
			(c.CommConfig.SSHPassword != "" || c.CommConfig.SSHPrivateKeyFile != "")) ||
			(c.CommConfig.SSHPassword != "" && c.CommConfig.SSHPrivateKeyFile != "") {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("only one of ssh_agent_auth, ssh_password, and ssh_private_key_file must be specified"))

		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
