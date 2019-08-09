package bsusurrogate

import (
	"context"
	"fmt"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepRegisterOMI creates the OMI.
type StepRegisterOMI struct {
	RootDevice    RootBlockDevice
	OMIDevices    []oapi.BlockDeviceMappingImage
	LaunchDevices []oapi.BlockDeviceMappingVmCreation
	image         *oapi.Image
}

func (s *StepRegisterOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oapiconn := state.Get("oapi").(*oapi.Client)
	snapshotIds := state.Get("snapshot_ids").(map[string]string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Registering the OMI...")

	blockDevices := s.combineDevices(snapshotIds)

	registerOpts := oapi.CreateImageRequest{
		ImageName:           config.OMIName,
		Architecture:        "x86_64",
		RootDeviceName:      s.RootDevice.DeviceName,
		BlockDeviceMappings: blockDevices,
	}

	registerResp, err := oapiconn.POST_CreateImage(registerOpts)
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering OMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	// Set the OMI ID in the state
	ui.Say(fmt.Sprintf("OMI: %s", registerResp.OK.Image.ImageId))
	omis := make(map[string]string)
	omis[oapiconn.GetConfig().Region] = registerResp.OK.Image.ImageId
	state.Put("omis", omis)

	// Wait for the image to become ready
	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilImageAvailable(oapiconn, registerResp.OK.Image.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imagesResp, err := oapiconn.POST_ReadImages(oapi.ReadImagesRequest{
		Filters: oapi.FiltersImage{
			ImageIds: []string{registerResp.OK.Image.ImageId},
		},
	})

	if err != nil {
		err := fmt.Errorf("Error searching for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = &imagesResp.OK.Images[0]

	snapshots := make(map[string][]string)
	for _, blockDeviceMapping := range imagesResp.OK.Images[0].BlockDeviceMappings {
		if blockDeviceMapping.Bsu.SnapshotId != "" {
			snapshots[oapiconn.GetConfig().Region] = append(snapshots[oapiconn.GetConfig().Region], blockDeviceMapping.Bsu.SnapshotId)
		}
	}
	state.Put("snapshots", snapshots)

	return multistep.ActionContinue
}

func (s *StepRegisterOMI) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deregistering the OMI because cancellation or error...")
	deregisterOpts := oapi.DeleteImageRequest{ImageId: s.image.ImageId}
	if _, err := oapiconn.POST_DeleteImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering OMI, may still be around: %s", err))
		return
	}
}

func (s *StepRegisterOMI) combineDevices(snapshotIds map[string]string) []oapi.BlockDeviceMappingImage {
	devices := map[string]oapi.BlockDeviceMappingImage{}

	for _, device := range s.OMIDevices {
		devices[device.DeviceName] = device
	}

	// Devices in launch_block_device_mappings override any with
	// the same name in ami_block_device_mappings, except for the
	// one designated as the root device in ami_root_device
	for _, device := range s.LaunchDevices {
		snapshotId, ok := snapshotIds[device.DeviceName]
		if ok {
			device.Bsu.SnapshotId = snapshotId
		}
		if device.DeviceName == s.RootDevice.SourceDeviceName {
			device.DeviceName = s.RootDevice.DeviceName
		}
		devices[device.DeviceName] = copyToDeviceMappingImage(device)
	}

	blockDevices := []oapi.BlockDeviceMappingImage{}
	for _, device := range devices {
		blockDevices = append(blockDevices, device)
	}
	return blockDevices
}

func copyToDeviceMappingImage(device oapi.BlockDeviceMappingVmCreation) oapi.BlockDeviceMappingImage {
	deviceImage := oapi.BlockDeviceMappingImage{
		DeviceName:        device.DeviceName,
		VirtualDeviceName: device.VirtualDeviceName,
		Bsu: oapi.BsuToCreate{
			DeleteOnVmDeletion: device.Bsu.DeleteOnVmDeletion,
			Iops:               device.Bsu.Iops,
			SnapshotId:         device.Bsu.SnapshotId,
			VolumeSize:         device.Bsu.VolumeSize,
			VolumeType:         device.Bsu.VolumeType,
		},
	}
	return deviceImage
}
