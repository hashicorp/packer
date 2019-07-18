package bsuvolume

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

type stepTagBSUVolumes struct {
	VolumeMapping []BlockDevice
	Ctx           interpolate.Context
}

func (s *stepTagBSUVolumes) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	vm := state.Get("vm").(oapi.Vm)
	ui := state.Get("ui").(packer.Ui)

	volumes := make(BsuVolumes)
	for _, instanceBlockDevices := range vm.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			if configVolumeMapping.DeviceName == instanceBlockDevices.DeviceName {
				volumes[oapiconn.GetConfig().Region] = append(
					volumes[oapiconn.GetConfig().Region],
					instanceBlockDevices.Bsu.VolumeId)
			}
		}
	}
	state.Put("bsuvolumes", volumes)

	if len(s.VolumeMapping) > 0 {
		ui.Say("Tagging BSU volumes...")

		toTag := map[string][]oapi.ResourceTag{}
		for _, mapping := range s.VolumeMapping {
			if len(mapping.Tags) == 0 {
				ui.Say(fmt.Sprintf("No tags specified for volume on %s...", mapping.DeviceName))
				continue
			}

			tags, err := mapping.Tags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
			if err != nil {
				err := fmt.Errorf("Error tagging device %s with %s", mapping.DeviceName, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			tags.Report(ui)

			for _, v := range vm.BlockDeviceMappings {
				if v.DeviceName == mapping.DeviceName {
					toTag[v.Bsu.VolumeId] = tags
				}
			}
		}

		for volumeId, tags := range toTag {
			_, err := oapiconn.POST_CreateTags(oapi.CreateTagsRequest{
				ResourceIds: []string{volumeId},
				Tags:        tags,
			})
			if err != nil {
				err := fmt.Errorf("Error tagging BSU Volume %s on %s: %s", volumeId, vm.VmId, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		}
	}

	return multistep.ActionContinue
}

func (s *stepTagBSUVolumes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
