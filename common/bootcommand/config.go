package bootcommand

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

type BootConfig struct {
	RawBootGroupInterval string        `mapstructure:"boot_keygroup_interval"`
	RawBootWait          string        `mapstructure:"boot_wait"`
	BootCommand          []string      `mapstructure:"boot_command"`
	BootGroupInterval    time.Duration ``
	BootWait             time.Duration ``
}

type VNCConfig struct {
	BootConfig `mapstructure:",squash"`
	DisableVNC bool `mapstructure:"disable_vnc"`
	// time in ms to wait between each key press
	RawBootKeyInterval string        `mapstructure:"boot_key_interval"`
	BootKeyInterval    time.Duration ``
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

	if c.RawBootGroupInterval == "" {
		c.RawBootGroupInterval = "0ms"
	}

	if c.RawBootGroupInterval != "" {
		bgi, err := time.ParseDuration(c.RawBootGroupInterval)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_keygroup_interval: %s", err))
		} else {
			c.BootGroupInterval = bgi
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

	if c.RawBootKeyInterval == "" {
		c.RawBootKeyInterval = "0ms"
	}

	if c.RawBootKeyInterval != "" {
		bki, err := time.ParseDuration(c.RawBootKeyInterval)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_key_interval: %s", err))
		} else {
			c.BootKeyInterval = bki
		}
	}

	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}
