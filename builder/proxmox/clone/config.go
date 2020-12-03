//go:generate mapstructure-to-hcl2 -type Config

package proxmoxclone

import (
	proxmox "github.com/hashicorp/packer/builder/proxmox/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
)

type Config struct {
	proxmox.Config `mapstructure:",squash"`

	CloneVM   string         `mapstructure:"clone_vm"`
	FullClone config.Trilean `mapstructure:"full_clone" required:"false"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, []string, error) {
	var errs *packersdk.MultiError
	_, warnings, merrs := c.Config.Prepare(c, raws...)
	if merrs != nil {
		errs = packersdk.MultiErrorAppend(errs, merrs)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}
	return nil, warnings, nil
}
