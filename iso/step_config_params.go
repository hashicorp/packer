package iso

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"fmt"
	"context"
)

type ConfigParamsConfig struct {
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`
}

func (c *ConfigParamsConfig) Prepare() []error {
	return nil
}

type StepConfigParams struct {
	Config *ConfigParamsConfig
}

func (s *StepConfigParams) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Adding configuration parameters...")

	if err := vm.AddConfigParams(s.Config.ConfigParams); err != nil {
		state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}
