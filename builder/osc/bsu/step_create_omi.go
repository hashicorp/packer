package bsu

import (
	"context"
	"fmt"
	"log"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

type stepCreateOMI struct {
	image *oapi.Image
}

func (s *stepCreateOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oapiconn := state.Get("oapi").(*oapi.Client)
	vm := state.Get("vm").(oapi.Vm)
	ui := state.Get("ui").(packer.Ui)

	// Create the image
	omiName := config.OMIName

	ui.Say(fmt.Sprintf("Creating OMI %s from vm %s", omiName, vm.VmId))
	createOpts := oapi.CreateImageRequest{
		VmId:                vm.VmId,
		ImageName:           omiName,
		BlockDeviceMappings: config.BlockDevices.BuildOMIDevices(),
	}

	resp, err := oapiconn.POST_CreateImage(createOpts)
	if err != nil || resp.OK == nil {
		err := fmt.Errorf("Error creating OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := resp.OK.Image

	// Set the OMI ID in the state
	ui.Message(fmt.Sprintf("OMI: %s", image.ImageId))
	omis := make(map[string]string)
	omis[oapiconn.GetConfig().Region] = image.ImageId
	state.Put("omis", omis)

	// Wait for the image to become ready
	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilImageAvailable(oapiconn, image.ImageId); err != nil {
		log.Printf("Error waiting for OMI: %s", err)
		imagesResp, err := oapiconn.POST_ReadImages(oapi.ReadImagesRequest{
			Filters: oapi.FiltersImage{
				ImageIds: []string{image.ImageId},
			},
		})
		if err != nil {
			log.Printf("Unable to determine reason waiting for OMI failed: %s", err)
			err = fmt.Errorf("Unknown error waiting for OMI.")
		} else {
			stateReason := imagesResp.OK.Images[0].StateComment
			err = fmt.Errorf("Error waiting for OMI. Reason: %s", stateReason)
		}

		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imagesResp, err := oapiconn.POST_ReadImages(oapi.ReadImagesRequest{
		Filters: oapi.FiltersImage{
			ImageIds: []string{image.ImageId},
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

func (s *stepCreateOMI) Cleanup(state multistep.StateBag) {
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
	DeleteOpts := oapi.DeleteImageRequest{ImageId: s.image.ImageId}
	if _, err := oapiconn.POST_DeleteImage(DeleteOpts); err != nil {
		ui.Error(fmt.Sprintf("Error Deleting OMI, may still be around: %s", err))
		return
	}
}
