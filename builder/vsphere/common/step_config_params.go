//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ConfigParamsConfig

package common

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type ConfigParamsConfig struct {
	// configuration_parameters is a direct passthrough to the VSphere API's
	// ConfigSpec: https://pubs.vmware.com/vi3/sdk/ReferenceGuide/vim.vm.ConfigSpec.html
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`

	// Enables time synchronization with the host. Defaults to false.
	ToolsSyncTime bool `mapstructure:"tools_sync_time"`

	// If sets to true, vSphere will automatically check and upgrade VMware Tools upon a system power cycle.
	// If not set, defaults to manual upgrade.
	ToolsUpgradePolicy bool `mapstructure:"tools_upgrade_policy"`
}

type StepConfigParams struct {
	Config *ConfigParamsConfig
}

func (s *StepConfigParams) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	configParams := make(map[string]string)

	if s.Config.ConfigParams != nil {
		configParams = s.Config.ConfigParams
	}

	var info *types.ToolsConfigInfo
	if s.Config.ToolsSyncTime || s.Config.ToolsUpgradePolicy {
		info = &types.ToolsConfigInfo{}

		if s.Config.ToolsSyncTime {
			info.SyncTimeWithHost = &s.Config.ToolsSyncTime
		}

		if s.Config.ToolsUpgradePolicy {
			info.ToolsUpgradePolicy = "UpgradeAtPowerCycle"
		}
	}

	ui.Say("Adding configuration parameters...")
	if err := vm.AddConfigParams(configParams, info); err != nil {
		state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}
