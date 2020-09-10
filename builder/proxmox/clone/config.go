//go:generate mapstructure-to-hcl2 -type Config

package proxmoxclone

import (
	"github.com/hashicorp/packer/builder/proxmox/common"
	"github.com/hashicorp/packer/packer"
)

type Config struct {
	proxmox.Config `mapstructure:",squash"`

	CloneVM string `mapstructure:"clone_vm"`
        FullClone bool `mapstructure:"full_clone"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, []string, error) {
	var errs *packer.MultiError
	_, warnings, merrs := c.Config.Prepare(c, raws...)
	if merrs != nil {
		errs = packer.MultiErrorAppend(errs, merrs)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}
	return nil, warnings, nil
}
