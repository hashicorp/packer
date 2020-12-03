package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/packer/builder/osc/common/retry"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/outscale/osc-sdk-go/osc"
)

type StepCreateTags struct {
	Tags         TagMap
	SnapshotTags TagMap
	Ctx          interpolate.Context
}

func (s *StepCreateTags) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("accessConfig").(*AccessConfig)
	ui := state.Get("ui").(packersdk.Ui)
	omis := state.Get("omis").(map[string]string)

	if !s.Tags.IsSet() && !s.SnapshotTags.IsSet() {
		return multistep.ActionContinue
	}

	// Adds tags to OMIs and snapshots
	for region, ami := range omis {
		ui.Say(fmt.Sprintf("Adding tags to OMI (%s)...", ami))

		regionconn := config.NewOSCClientByRegion(region)

		// Retrieve image list for given OMI
		resourceIds := []string{ami}
		imageResp, _, err := regionconn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
			ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
				Filters: osc.FiltersImage{
					ImageIds: resourceIds,
				},
			}),
		})

		if err != nil {
			err := fmt.Errorf("Error retrieving details for OMI (%s): %s", ami, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(imageResp.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for OMI (%s), no images found", ami)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		image := imageResp.Images[0]
		snapshotIds := []string{}

		// Add only those with a Snapshot ID, i.e. not Ephemeral
		for _, device := range image.BlockDeviceMappings {
			if device.Bsu.SnapshotId != "" {
				ui.Say(fmt.Sprintf("Tagging snapshot: %s", device.Bsu.SnapshotId))
				resourceIds = append(resourceIds, device.Bsu.SnapshotId)
				snapshotIds = append(snapshotIds, device.Bsu.SnapshotId)
			}
		}

		// Convert tags to oapi.Tag format
		ui.Say("Creating OMI tags")
		amiTags, err := s.Tags.OSCTags(s.Ctx, config.RawRegion, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		amiTags.Report(ui)

		ui.Say("Creating snapshot tags")
		snapshotTags, err := s.SnapshotTags.OSCTags(s.Ctx, config.RawRegion, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		snapshotTags.Report(ui)

		// Retry creating tags for about 2.5 minutes
		err = retry.Run(0.2, 30, 11, func(_ uint) (bool, error) {
			// Tag images and snapshots
			_, _, err := regionconn.TagApi.CreateTags(context.Background(), &osc.CreateTagsOpts{
				CreateTagsRequest: optional.NewInterface(osc.CreateTagsRequest{
					ResourceIds: resourceIds,
					Tags:        amiTags,
				}),
			})
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidOMIID.NotFound" ||
					awsErr.Code() == "InvalidSnapshot.NotFound" {
					return false, nil
				}
			}

			// Override tags on snapshots
			if len(snapshotTags) > 0 {
				_, _, err = regionconn.TagApi.CreateTags(context.Background(), &osc.CreateTagsOpts{
					CreateTagsRequest: optional.NewInterface(osc.CreateTagsRequest{
						ResourceIds: snapshotIds,
						Tags:        snapshotTags,
					}),
				})
			}
			if err == nil {
				return true, nil
			}
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidSnapshot.NotFound" {
					return false, nil
				}
			}
			return true, err
		})

		if err != nil {
			err := fmt.Errorf("Error adding tags to Resources (%#v): %s", resourceIds, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
