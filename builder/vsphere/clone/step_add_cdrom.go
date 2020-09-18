//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type CDRomConfig

package clone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type CDRomConfig struct {
	// Which controller to use. Example: `sata`. Defaults to `ide`.
	CdromType string `mapstructure:"cdrom_type"`
}

type StepAddCDRom struct {
	Config *CDRomConfig
}

func (c *CDRomConfig) Prepare() []error {
	var errs []error

	if c.CdromType != "" && c.CdromType != "ide" && c.CdromType != "sata" {
		errs = append(errs, fmt.Errorf("'cdrom_type' must be 'ide' or 'sata'"))
	}

	return errs
}

func (s *StepAddCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	if s.Config.CdromType == "sata" {
		if _, err := vm.FindSATAController(); err == driver.ErrNoSataController {
			ui.Say("Adding SATA controller...")
			if err := vm.AddSATAController(); err != nil {
				state.Put("error", fmt.Errorf("error adding SATA controller: %v", err))
				return multistep.ActionHalt
			}
		}
	}

	// Add our custom CD, if it exists
	if cd_path, _ := state.Get("cd_path").(string); cd_path != "" {
		if err := vm.AddCdrom(s.Config.CdromType, cd_path); err != nil {
			state.Put("error", fmt.Errorf("error mounting a CD '%v': %v", cd_path, err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {}
