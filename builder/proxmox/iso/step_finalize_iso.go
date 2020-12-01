package proxmoxiso

import (
	"context"
	"fmt"
	"strings"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// stepFinalizeISOTemplate does any ISO-builder specific modifications after
// conversion to a template, and after the non-specific modifications in
// common.stepFinalizeTemplateConfig
type stepFinalizeISOTemplate struct{}

type templateFinalizer interface {
	GetVmConfig(*proxmox.VmRef) (map[string]interface{}, error)
	SetVmConfig(*proxmox.VmRef, map[string]interface{}) (interface{}, error)
}

var _ templateFinalizer = &proxmox.Client{}

func (s *stepFinalizeISOTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("proxmoxClient").(templateFinalizer)
	c := state.Get("iso-config").(*Config)
	vmRef := state.Get("vmRef").(*proxmox.VmRef)

	changes := make(map[string]interface{})

	if c.UnmountISO {
		vmParams, err := client.GetVmConfig(vmRef)
		if err != nil {
			err := fmt.Errorf("Error fetching template config: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if vmParams["ide2"] == nil || !strings.HasSuffix(vmParams["ide2"].(string), "media=cdrom") {
			err := fmt.Errorf("Cannot eject ISO from cdrom drive, ide2 is not present, or not a cdrom media")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		changes["ide2"] = "none,media=cdrom"
	}
	if len(c.AdditionalISOFiles) > 0 {
		vmParams, err := client.GetVmConfig(vmRef)
		if err != nil {
			err := fmt.Errorf("Error fetching template config: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		for idx := range c.AdditionalISOFiles {
			cdrom := c.AdditionalISOFiles[idx].Device
			if c.AdditionalISOFiles[idx].Unmount {
				if vmParams[cdrom] == nil || !strings.Contains(vmParams[cdrom].(string), "media=cdrom") {
					err := fmt.Errorf("Cannot eject ISO from cdrom drive, %s is not present or not a cdrom media", cdrom)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				changes[cdrom] = "none,media=cdrom"
			} else {
				changes[cdrom] = c.AdditionalISOFiles[idx].ISOFile + ",media=cdrom"
			}
		}
	}

	if len(changes) > 0 {
		_, err := client.SetVmConfig(vmRef, changes)
		if err != nil {
			err := fmt.Errorf("Error updating template: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepFinalizeISOTemplate) Cleanup(state multistep.StateBag) {
}
