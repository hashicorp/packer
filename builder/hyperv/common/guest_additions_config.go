package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

type GuestAdditionsConfig struct {
	// Should integration services iso be mounted
	GuestAdditionsMode string `mapstructure:"guest_additions_mode"`

	// The path to the integration services iso
	GuestAdditionsPath string `mapstructure:"guest_additions_path"`
}

func (c *GuestAdditionsConfig) Prepare(ctx *interpolate.Context, numberOfIsos int, generation uint) (errs []error) {
	// Errors
	if c.GuestAdditionsMode == "" {
		if c.GuestAdditionsPath != "" {
			c.GuestAdditionsMode = "attach"
		} else {
			c.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

			if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
				if err != nil {
					c.GuestAdditionsPath = ""
					c.GuestAdditionsMode = "none"
				} else {
					c.GuestAdditionsMode = "attach"
				}
			}
		}
	}

	if c.GuestAdditionsPath == "" && c.GuestAdditionsMode == "attach" {
		c.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

		if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				c.GuestAdditionsPath = ""
			}
		}
	}

	totalNumIsos := numberOfIsos
	if c.GuestAdditionsMode == "attach" {
		if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Guest additions iso does not exist: %s", err))
			}
		}

		totalNumIsos++
	}

	if generation < 2 && totalNumIsos > 2 {
		if c.GuestAdditionsMode == "attach" {
			errs = append(errs, fmt.Errorf("There are only 2 ide controllers available, so we can't support guest additions and these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = append(errs, fmt.Errorf("There are only 2 ide controllers available, so we can't support these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	} else if generation > 1 && numberOfIsos > 16 {
		if c.GuestAdditionsMode == "attach" {
			errs = append(errs, fmt.Errorf("There are not enough drive letters available for scsi (limited to 16), so we can't support guest additions and these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = append(errs, fmt.Errorf("There are not enough drive letters available for scsi (limited to 16), so we can't support these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	}

	return
}
