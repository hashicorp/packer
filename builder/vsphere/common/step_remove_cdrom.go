//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type RemoveCDRomConfig

package common

import (
	"context"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type RemoveCDRomConfig struct {
	// Remove CD-ROM devices from template. Defaults to `false`.
	RemoveCdrom bool `mapstructure:"remove_cdrom"`
}

type StepRemoveCDRom struct {
	Config *RemoveCDRomConfig
}

func (s *StepRemoveCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)

	ui.Say("Eject CD-ROM drives...")
	err := vm.EjectCdroms()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if s.Config.RemoveCdrom == true {
		ui.Say("Deleting CD-ROM drives...")
		err := vm.RemoveCdroms()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveCDRom) Cleanup(state multistep.StateBag) {}
