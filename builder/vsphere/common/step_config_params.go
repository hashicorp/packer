//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ConfigParamsConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type ConfigParamsConfig struct {
	// configuration_parameters is a direct passthrough to the VSphere API's
	// ConfigSpec: https://pubs.vmware.com/vi3/sdk/ReferenceGuide/vim.vm.ConfigSpec.html
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`

	// Enables time synchronization with the host. If set to true will set `tools.syncTime` to `TRUE`.
	// Defaults to FALSE.
	ToolsSyncTime bool `mapstructure:"tools_sync_time"`

	// If sets to true, vSphere will automatically check and upgrade VMware Tools upon a system power cycle.
	// If not set, defaults to manual upgrade.
	ToolsUpgradePolicy bool `mapstructure:"tools_upgrade_policy"`
}

type StepConfigParams struct {
	Config *ConfigParamsConfig
}

func (s *StepConfigParams) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	configParams := make(map[string]string)

	if s.Config.ConfigParams != nil {
		configParams = s.Config.ConfigParams
	}

	if s.Config.ToolsSyncTime {
		configParams["tools.syncTime"] = "TRUE"
	}

	if s.Config.ToolsUpgradePolicy {
		configParams["tools.upgrade.policy"] = "upgradeAtPowerCycle"
	}

	if len(configParams) > 0 {
		ui.Say("Adding configuration parameters...")
		if err := vm.AddConfigParams(configParams); err != nil {
			state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}
