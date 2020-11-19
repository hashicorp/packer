package chroot

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepCreateOMI creates the OMI.
type StepCreateOMI struct {
	RootVolumeSize int64
	RawRegion      string
}

func (s *StepCreateOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	osconn := state.Get("osc").(*osc.APIClient)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating the OMI...")

	var (
		registerOpts   osc.CreateImageRequest
		mappings       []osc.BlockDeviceMappingImage
		image          osc.Image
		rootDeviceName string
	)

	if config.FromScratch {
		mappings = config.OMIBlockDevices.BuildOscOMIDevices()
		rootDeviceName = config.RootDeviceName
	} else {
		image = state.Get("source_image").(osc.Image)
		mappings = image.BlockDeviceMappings
		rootDeviceName = image.RootDeviceName
	}

	newMappings := make([]osc.BlockDeviceMappingImage, len(mappings))
	for i, device := range mappings {
		newDevice := device

		//FIX: Temporary fix
		gibSize := newDevice.Bsu.VolumeSize / (1024 * 1024 * 1024)
		newDevice.Bsu.VolumeSize = gibSize

		if newDevice.DeviceName == rootDeviceName {
			if newDevice.Bsu != (osc.BsuToCreate{}) {
				newDevice.Bsu.SnapshotId = snapshotId
			} else {
				newDevice.Bsu = osc.BsuToCreate{SnapshotId: snapshotId}
			}

			if config.FromScratch || int32(s.RootVolumeSize) > newDevice.Bsu.VolumeSize {
				newDevice.Bsu.VolumeSize = int32(s.RootVolumeSize)
			}
		}

		newMappings[i] = newDevice
	}

	if config.FromScratch {
		registerOpts = osc.CreateImageRequest{
			ImageName:           config.OMIName,
			Architecture:        "x86_64",
			RootDeviceName:      rootDeviceName,
			BlockDeviceMappings: newMappings,
		}
	} else {
		registerOpts = buildRegisterOpts(config, image, newMappings)
	}

	registerResp, _, err := osconn.ImageApi.CreateImage(context.Background(), &osc.CreateImageOpts{
		CreateImageRequest: optional.NewInterface(registerOpts),
	})
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering OMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	imageID := registerResp.Image.ImageId

	// Set the OMI ID in the state
	ui.Say(fmt.Sprintf("OMI: %s", imageID))
	omis := make(map[string]string)
	omis[s.RawRegion] = imageID
	state.Put("omis", omis)

	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilOscImageAvailable(osconn, imageID); err != nil {
		err := fmt.Errorf("Error waiting for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateOMI) Cleanup(state multistep.StateBag) {}

func buildRegisterOpts(config *Config, image osc.Image, mappings []osc.BlockDeviceMappingImage) osc.CreateImageRequest {
	registerOpts := osc.CreateImageRequest{
		ImageName:           config.OMIName,
		Architecture:        image.Architecture,
		RootDeviceName:      image.RootDeviceName,
		BlockDeviceMappings: mappings,
	}
	return registerOpts
}
