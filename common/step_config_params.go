package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type ConfigParamsConfig struct {
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`
}

type StepConfigParams struct {
	Config *ConfigParamsConfig
}

func (s *StepConfigParams) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if s.Config.ConfigParams != nil {
		ui.Say("Adding configuration parameters...")
		if err := vm.AddConfigParams(s.Config.ConfigParams); err != nil {
			state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}
