package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

// StepCreateVolume creates a new volume from the snapshot of the root
// device of the OMI.
//
// Produces:
//   volume_id string - The ID of the created volume
type StepCreateVolume struct {
	volumeId       string
	RootVolumeSize int64
	RootVolumeType string
	RootVolumeTags osccommon.TagMap
	Ctx            interpolate.Context
}

func (s *StepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oapiconn := state.Get("oapi").(*oapi.Client)
	vm := state.Get("vm").(oapi.Vm)
	ui := state.Get("ui").(packer.Ui)

	var err error

	//TODO: Add tags
	// volTags, err := s.RootVolumeTags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
	// if err != nil {
	// 	err := fmt.Errorf("Error tagging volumes: %s", err)
	// 	state.Put("error", err)
	// 	ui.Error(err.Error())
	// 	return multistep.ActionHalt
	// }

	// // Collect tags for tagging on resource creation
	// var tagSpecs []oapi.ResourceTag

	// if len(volTags) > 0 {
	// 	runVolTags := &oapi.Resou rceTag{
	// 		ResourceType: "volume",
	// 		Tags:         volTags,
	// 	}

	// 	tagSpecs = append(tagSpecs, runVolTags)
	// }

	var createVolume *oapi.CreateVolumeRequest
	if config.FromScratch {
		rootVolumeType := osccommon.VolumeTypeGp2
		if s.RootVolumeType == "io1" {
			err := errors.New("Cannot use io1 volume when building from scratch")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else if s.RootVolumeType != "" {
			rootVolumeType = s.RootVolumeType
		}
		createVolume = &oapi.CreateVolumeRequest{
			SubregionName: vm.Placement.SubregionName,
			Size:          s.RootVolumeSize,
			VolumeType:    rootVolumeType,
		}

	} else {
		// Determine the root device snapshot
		image := state.Get("source_image").(oapi.Image)
		log.Printf("Searching for root device of the image (%s)", image.RootDeviceName)
		var rootDevice *oapi.BlockDeviceMappingImage
		for _, device := range image.BlockDeviceMappings {
			if device.DeviceName == image.RootDeviceName {
				rootDevice = &device
				break
			}
		}

		ui.Say("Creating the root volume...")
		createVolume, err = s.buildCreateVolumeInput(vm.Placement.SubregionName, rootDevice)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	//TODO: ADD TAGS
	// if len(tagSpecs) > 0 {
	// 	createVolume.SetTagSpecifications(tagSpecs)
	// 	volTags.Report(ui)
	// }
	log.Printf("Create args: %+v", createVolume)

	createVolumeResp, err := oapiconn.POST_CreateVolume(*createVolume)
	if err != nil {
		err := fmt.Errorf("Error creating root volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the volume ID so we remember to delete it later
	s.volumeId = createVolumeResp.OK.Volume.VolumeId
	log.Printf("Volume ID: %s", s.volumeId)

	// Wait for the volume to become ready
	err = osccommon.WaitUntilVolumeAvailable(oapiconn, s.volumeId)
	if err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("volume_id", s.volumeId)
	return multistep.ActionContinue
}

func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	if s.volumeId == "" {
		return
	}

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting the created BSU volume...")
	_, err := oapiconn.POST_DeleteVolume(oapi.DeleteVolumeRequest{VolumeId: s.volumeId})
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting BSU volume: %s", err))
	}
}

func (s *StepCreateVolume) buildCreateVolumeInput(suregionName string, rootDevice *oapi.BlockDeviceMappingImage) (*oapi.CreateVolumeRequest, error) {
	if rootDevice == nil {
		return nil, fmt.Errorf("Couldn't find root device!")
	}

	//FIX: Temporary fix
	gibSize := rootDevice.Bsu.VolumeSize / (1024 * 1024 * 1024)
	createVolumeInput := &oapi.CreateVolumeRequest{
		SubregionName: suregionName,
		Size:          gibSize,
		SnapshotId:    rootDevice.Bsu.SnapshotId,
		VolumeType:    rootDevice.Bsu.VolumeType,
		Iops:          rootDevice.Bsu.Iops,
	}
	if s.RootVolumeSize > rootDevice.Bsu.VolumeSize {
		createVolumeInput.Size = s.RootVolumeSize
	}

	if s.RootVolumeType == "" || s.RootVolumeType == rootDevice.Bsu.VolumeType {
		return createVolumeInput, nil
	}

	if s.RootVolumeType == "io1" {
		return nil, fmt.Errorf("Root volume type cannot be io1, because existing root volume type was %s", rootDevice.Bsu.VolumeType)
	}

	createVolumeInput.VolumeType = s.RootVolumeType
	// non io1 cannot set iops
	createVolumeInput.Iops = 0

	return createVolumeInput, nil
}
