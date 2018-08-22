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
	// time in ms to wait between each group of 25 key presses
	BootGroupInterval int `mapstructure:"boot_keygroup_interval"`

	BootWait time.Duration ``
}

type VNCConfig struct {
	BootConfig `mapstructure:",squash"`
	DisableVNC bool `mapstructure:"disable_vnc"`
	// time in ms to wait between each key press
	BootKeyInterval int `mapstructure:"boot_key_interval"`
}

func (c *BootConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}
	if c.BootGroupInterval == 0 {
		c.BootGroupInterval = -1
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

	if c.BootCommand != nil {
		expSeq, err := GenerateExpressionSequence(c.FlatBootCommand())
		if err != nil {
			errs = append(errs, err)
		} else if vErrs := expSeq.Validate(); vErrs != nil {
			errs = append(errs, vErrs...)
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
	if c.BootKeyInterval == 0 {
		c.BootKeyInterval = -1
	}
	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}
