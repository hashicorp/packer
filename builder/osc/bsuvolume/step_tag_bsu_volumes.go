package bsuvolume

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/outscale/osc-sdk-go/osc"
)

type stepTagBSUVolumes struct {
	VolumeMapping []BlockDevice
	RawRegion     string
	Ctx           interpolate.Context
}

func (s *stepTagBSUVolumes) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	vm := state.Get("vm").(osc.Vm)
	ui := state.Get("ui").(packersdk.Ui)

	volumes := make(BsuVolumes)
	for _, instanceBlockDevices := range vm.BlockDeviceMappings {
		for _, configVolumeMapping := range s.VolumeMapping {
			if configVolumeMapping.DeviceName == instanceBlockDevices.DeviceName {
				volumes[s.RawRegion] = append(
					volumes[s.RawRegion],
					instanceBlockDevices.Bsu.VolumeId)
			}
		}
	}
	state.Put("bsuvolumes", volumes)

	if len(s.VolumeMapping) > 0 {
		ui.Say("Tagging BSU volumes...")

		toTag := map[string][]osc.ResourceTag{}
		for _, mapping := range s.VolumeMapping {
			if len(mapping.Tags) == 0 {
				ui.Say(fmt.Sprintf("No tags specified for volume on %s...", mapping.DeviceName))
				continue
			}

			tags, err := mapping.Tags.OSCTags(s.Ctx, s.RawRegion, state)
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
			_, _, err := oscconn.TagApi.CreateTags(context.Background(), &osc.CreateTagsOpts{
				CreateTagsRequest: optional.NewInterface(osc.CreateTagsRequest{
					ResourceIds: []string{volumeId},
					Tags:        tags,
				}),
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
