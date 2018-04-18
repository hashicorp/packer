package bootcommand

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

type BootConfig struct {
	RawBootWait string   `mapstructure:"boot_wait"`
	BootCommand []string `mapstructure:"boot_command"`

	BootWait time.Duration ``
}

type VNCConfig struct {
	BootConfig
	DisableVNC bool `mapstructure:"disable_vnc"`
}

func (c *BootConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}
	if c.RawBootWait != "" {
		bw, err := time.ParseDuration(c.RawBootWait)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		} else {
			c.BootWait = bw
		}
	}
	return
}

func (c *BootConfig) FlatBootCommand() string {
	return strings.Join(c.BootCommand, "")
}

func (c *VNCConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if len(c.BootCommand) > 0 && c.DisableVNC {
		errs = append(errs,
			fmt.Errorf("A boot command cannot be used when vnc is disabled."))
	}
	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}
