package chroot

import (
	"context"
	"fmt"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepCreateOMI creates the OMI.
type StepCreateOMI struct {
	RootVolumeSize int64
}

func (s *StepCreateOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oapiconn := state.Get("oapi").(*oapi.Client)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating the OMI...")

	var (
		registerOpts   oapi.CreateImageRequest
		mappings       []oapi.BlockDeviceMappingImage
		image          oapi.Image
		rootDeviceName string
	)

	if config.FromScratch {
		mappings = config.OMIBlockDevices.BuildOMIDevices()
		rootDeviceName = config.RootDeviceName
	} else {
		image = state.Get("source_image").(oapi.Image)
		mappings = image.BlockDeviceMappings
		rootDeviceName = image.RootDeviceName
	}

	newMappings := make([]oapi.BlockDeviceMappingImage, len(mappings))
	for i, device := range mappings {
		newDevice := device

		//FIX: Temporary fix
		gibSize := newDevice.Bsu.VolumeSize / (1024 * 1024 * 1024)
		newDevice.Bsu.VolumeSize = gibSize

		if newDevice.DeviceName == rootDeviceName {
			if newDevice.Bsu != (oapi.BsuToCreate{}) {
				newDevice.Bsu.SnapshotId = snapshotId
			} else {
				newDevice.Bsu = oapi.BsuToCreate{SnapshotId: snapshotId}
			}

			if config.FromScratch || s.RootVolumeSize > newDevice.Bsu.VolumeSize {
				newDevice.Bsu.VolumeSize = s.RootVolumeSize
			}
		}

		newMappings[i] = newDevice
	}

	if config.FromScratch {
		registerOpts = oapi.CreateImageRequest{
			ImageName:           config.OMIName,
			Architecture:        "x86_64",
			RootDeviceName:      rootDeviceName,
			BlockDeviceMappings: newMappings,
		}
	} else {
		registerOpts = buildRegisterOpts(config, image, newMappings)
	}

	registerResp, err := oapiconn.POST_CreateImage(registerOpts)
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering OMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	imageID := registerResp.OK.Image.ImageId

	// Set the OMI ID in the state
	ui.Say(fmt.Sprintf("OMI: %s", imageID))
	omis := make(map[string]string)
	omis[oapiconn.GetConfig().Region] = imageID
	state.Put("omis", omis)

	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilImageAvailable(oapiconn, imageID); err != nil {
		err := fmt.Errorf("Error waiting for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateOMI) Cleanup(state multistep.StateBag) {}

func buildRegisterOpts(config *Config, image oapi.Image, mappings []oapi.BlockDeviceMappingImage) oapi.CreateImageRequest {
	registerOpts := oapi.CreateImageRequest{
		ImageName:           config.OMIName,
		Architecture:        image.Architecture,
		RootDeviceName:      image.RootDeviceName,
		BlockDeviceMappings: mappings,
	}
	return registerOpts
}
