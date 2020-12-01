package bsu

import (
	"context"
	"fmt"
	"log"

	"github.com/antihax/optional"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

type stepCreateOMI struct {
	image     *osc.Image
	RawRegion string
}

func (s *stepCreateOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oscconn := state.Get("osc").(*osc.APIClient)
	vm := state.Get("vm").(osc.Vm)
	ui := state.Get("ui").(packersdk.Ui)

	// Create the image
	omiName := config.OMIName

	ui.Say(fmt.Sprintf("Creating OMI %s from vm %s", omiName, vm.VmId))
	createOpts := osc.CreateImageRequest{
		VmId:                vm.VmId,
		ImageName:           omiName,
		BlockDeviceMappings: config.BlockDevices.BuildOscOMIDevices(),
	}

	resp, _, err := oscconn.ImageApi.CreateImage(context.Background(), &osc.CreateImageOpts{
		CreateImageRequest: optional.NewInterface(createOpts),
	})
	if err != nil || resp.Image.ImageId == "" {
		err := fmt.Errorf("Error creating OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := resp.Image

	// Set the OMI ID in the state
	ui.Message(fmt.Sprintf("OMI: %s", image.ImageId))
	omis := make(map[string]string)
	omis[s.RawRegion] = image.ImageId
	state.Put("omis", omis)

	// Wait for the image to become ready
	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilOscImageAvailable(oscconn, image.ImageId); err != nil {
		log.Printf("Error waiting for OMI: %s", err)
		imagesResp, _, err := oscconn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
			ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
				Filters: osc.FiltersImage{
					ImageIds: []string{image.ImageId},
				},
			}),
		})
		if err != nil {
			log.Printf("Unable to determine reason waiting for OMI failed: %s", err)
			err = fmt.Errorf("Unknown error waiting for OMI")
		} else {
			stateReason := imagesResp.Images[0].StateComment
			err = fmt.Errorf("Error waiting for OMI. Reason: %s", stateReason)
		}

		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imagesResp, _, err := oscconn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
		ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
			Filters: osc.FiltersImage{
				ImageIds: []string{image.ImageId},
			},
		}),
	})
	if err != nil {
		err := fmt.Errorf("Error searching for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = &imagesResp.Images[0]

	snapshots := make(map[string][]string)
	for _, blockDeviceMapping := range imagesResp.Images[0].BlockDeviceMappings {
		if blockDeviceMapping.Bsu.SnapshotId != "" {
			snapshots[s.RawRegion] = append(snapshots[s.RawRegion], blockDeviceMapping.Bsu.SnapshotId)
		}
	}
	state.Put("snapshots", snapshots)

	return multistep.ActionContinue
}

func (s *stepCreateOMI) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	oscconn := state.Get("osc").(*osc.APIClient)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deregistering the OMI because cancellation or error...")
	DeleteOpts := osc.DeleteImageRequest{ImageId: s.image.ImageId}
	if _, _, err := oscconn.ImageApi.DeleteImage(context.Background(), &osc.DeleteImageOpts{
		DeleteImageRequest: optional.NewInterface(DeleteOpts),
	}); err != nil {
		ui.Error(fmt.Sprintf("Error Deleting OMI, may still be around: %s", err))
		return
	}
}
