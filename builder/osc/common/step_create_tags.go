package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/awserr"
	retry "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

type StepCreateTags struct {
	Tags         TagMap
	SnapshotTags TagMap
	Ctx          interpolate.Context
}

func (s *StepCreateTags) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	config := state.Get("clientConfig").(*oapi.Config)
	ui := state.Get("ui").(packer.Ui)
	omis := state.Get("omis").(map[string]string)

	if !s.Tags.IsSet() && !s.SnapshotTags.IsSet() {
		return multistep.ActionContinue
	}

	// Adds tags to OMIs and snapshots
	for region, ami := range omis {
		ui.Say(fmt.Sprintf("Adding tags to OMI (%s)...", ami))

		newConfig := &oapi.Config{
			UserAgent: config.UserAgent,
			SecretKey: config.SecretKey,
			Service:   config.Service,
			Region:    region, //New region
			URL:       config.URL,
		}

		skipClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		regionConn := oapi.NewClient(newConfig, skipClient)

		// Retrieve image list for given OMI
		resourceIds := []string{ami}
		imageResp, err := regionConn.POST_ReadImages(oapi.ReadImagesRequest{
			Filters: oapi.FiltersImage{
				ImageIds: resourceIds,
			},
		})

		if err != nil {
			err := fmt.Errorf("Error retrieving details for OMI (%s): %s", ami, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(imageResp.OK.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for OMI (%s), no images found", ami)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		image := imageResp.OK.Images[0]
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
		amiTags, err := s.Tags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		amiTags.Report(ui)

		ui.Say("Creating snapshot tags")
		snapshotTags, err := s.SnapshotTags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		snapshotTags.Report(ui)

		// Retry creating tags for about 2.5 minutes
		err = retry.Retry(0.2, 30, 11, func(_ uint) (bool, error) {
			// Tag images and snapshots
			_, err := regionConn.POST_CreateTags(oapi.CreateTagsRequest{
				ResourceIds: resourceIds,
				Tags:        amiTags,
			})
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidOMIID.NotFound" ||
					awsErr.Code() == "InvalidSnapshot.NotFound" {
					return false, nil
				}
			}

			// Override tags on snapshots
			if len(snapshotTags) > 0 {
				_, err = regionConn.POST_CreateTags(oapi.CreateTagsRequest{
					ResourceIds: snapshotIds,
					Tags:        snapshotTags,
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
