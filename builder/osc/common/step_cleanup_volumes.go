package common

import (
	"context"
	"fmt"
	"reflect"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepCleanupVolumes cleans up any orphaned volumes that were not designated to
// remain after termination of the vm. These volumes are typically ones
// that are marked as "delete on terminate:false" in the source_ami of a build.
type StepCleanupVolumes struct {
	BlockDevices BlockDevices
}

//Run ...
func (s *StepCleanupVolumes) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// stepCleanupVolumes is for Cleanup only
	return multistep.ActionContinue
}

// Cleanup ...
func (s *StepCleanupVolumes) Cleanup(state multistep.StateBag) {
	oscconn := state.Get("osc").(*osc.APIClient)
	vmRaw := state.Get("vm")
	var vm osc.Vm
	if vmRaw != nil {
		vm = vmRaw.(osc.Vm)
	}
	ui := state.Get("ui").(packersdk.Ui)
	if vm.VmId == "" {
		ui.Say("No volumes to clean up, skipping")
		return
	}

	ui.Say("Cleaning up any extra volumes...")

	// Collect Volume information from the cached Vm as a map of volume-id
	// to device name, to compare with save list below
	var vl []string
	volList := make(map[string]string)
	for _, bdm := range vm.BlockDeviceMappings {
		if !reflect.DeepEqual(bdm.Bsu, osc.BsuCreated{}) {
			vl = append(vl, bdm.Bsu.VolumeId)
			volList[bdm.Bsu.VolumeId] = bdm.DeviceName
		}
	}

	// Using the volume list from the cached Vm, check with Outscale for up to
	// date information on them
	resp, _, err := oscconn.VolumeApi.ReadVolumes(context.Background(), &osc.ReadVolumesOpts{
		ReadVolumesRequest: optional.NewInterface(osc.ReadVolumesRequest{
			Filters: osc.FiltersVolume{
				VolumeIds: vl,
			},
		}),
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Error describing volumes: %s", err))
		return
	}

	// If any of the returned volumes are in a "deleting" stage or otherwise not
	// available, remove them from the list of volumes
	for _, v := range resp.Volumes {
		if v.State != "" && v.State != "available" {
			delete(volList, v.VolumeId)
		}
	}

	if len(resp.Volumes) == 0 {
		ui.Say("No volumes to clean up, skipping")
		return
	}

	// Filter out any devices created as part of the launch mappings, since
	// we'll let outscale follow the `delete_on_vm_deletion` setting.
	for _, b := range s.BlockDevices.LaunchMappings {
		for volKey, volName := range volList {
			if volName == b.DeviceName {
				delete(volList, volKey)
			}
		}
	}

	// Destroy remaining volumes
	for k := range volList {
		ui.Say(fmt.Sprintf("Destroying volume (%s)...", k))
		_, _, err := oscconn.VolumeApi.DeleteVolume(context.Background(), &osc.DeleteVolumeOpts{
			DeleteVolumeRequest: optional.NewInterface(osc.DeleteVolumeRequest{VolumeId: k}),
		})
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting volume: %s", err))
		}
	}
}
