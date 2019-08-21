package common

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// stepCleanupVolumes cleans up any orphaned volumes that were not designated to
// remain after termination of the vm. These volumes are typically ones
// that are marked as "delete on terminate:false" in the source_ami of a build.
type StepCleanupVolumes struct {
	BlockDevices BlockDevices
}

func (s *StepCleanupVolumes) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// stepCleanupVolumes is for Cleanup only
	return multistep.ActionContinue
}

func (s *StepCleanupVolumes) Cleanup(state multistep.StateBag) {
	oapiconn := state.Get("oapi").(*oapi.Client)
	vmRaw := state.Get("vm")
	var vm oapi.Vm
	if vmRaw != nil {
		vm = vmRaw.(oapi.Vm)
	}
	ui := state.Get("ui").(packer.Ui)
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
		if !reflect.DeepEqual(bdm.Bsu, oapi.BsuCreated{}) {
			vl = append(vl, bdm.Bsu.VolumeId)
			volList[bdm.Bsu.VolumeId] = bdm.DeviceName
		}
	}

	// Using the volume list from the cached Vm, check with Outscale for up to
	// date information on them
	resp, err := oapiconn.POST_ReadVolumes(oapi.ReadVolumesRequest{
		Filters: oapi.FiltersVolume{
			VolumeIds: vl,
		},
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Error describing volumes: %s", err))
		return
	}

	// If any of the returned volumes are in a "deleting" stage or otherwise not
	// available, remove them from the list of volumes
	for _, v := range resp.OK.Volumes {
		if v.State != "" && v.State != "available" {
			delete(volList, v.VolumeId)
		}
	}

	if len(resp.OK.Volumes) == 0 {
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
		_, err := oapiconn.POST_DeleteVolume(oapi.DeleteVolumeRequest{VolumeId: k})
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting volume: %s", err))
		}

	}
}
