package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

type StepDeregisterOMI struct {
	AccessConfig        *AccessConfig
	ForceDeregister     bool
	ForceDeleteSnapshot bool
	OMIName             string
	Regions             []string
}

func (s *StepDeregisterOMI) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Check for force deregister
	if !s.ForceDeregister {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	oapiconn := state.Get("oapi").(*oapi.Client)
	// Add the session region to list of regions will will deregister OMIs in
	regions := append(s.Regions, oapiconn.GetConfig().Region)

	for _, region := range regions {
		// get new connection for each region in which we need to deregister vms
		config, err := s.AccessConfig.Config()
		if err != nil {
			return multistep.ActionHalt
		}

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

		regionconn := oapi.NewClient(newConfig, skipClient)

		resp, err := regionconn.POST_ReadImages(oapi.ReadImagesRequest{
			Filters: oapi.FiltersImage{
				ImageNames:     []string{s.OMIName},
				AccountAliases: []string{"self"},
			},
		})

		if err != nil {
			err := fmt.Errorf("Error describing OMI: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Deregister image(s) by name
		for _, i := range resp.OK.Images {
			//We are supposing that DeleteImage does the same action as DeregisterImage
			_, err := regionconn.POST_DeleteImage(oapi.DeleteImageRequest{
				ImageId: i.ImageId,
			})

			if err != nil {
				err := fmt.Errorf("Error deregistering existing OMI: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("Deregistered OMI %s, id: %s", s.OMIName, i.ImageId))

			// Delete snapshot(s) by image
			if s.ForceDeleteSnapshot {
				for _, b := range i.BlockDeviceMappings {
					if b.Bsu.SnapshotId != "" {
						_, err := regionconn.POST_DeleteSnapshot(oapi.DeleteSnapshotRequest{
							SnapshotId: b.Bsu.SnapshotId,
						})

						if err != nil {
							err := fmt.Errorf("Error deleting existing snapshot: %s", err)
							state.Put("error", err)
							ui.Error(err.Error())
							return multistep.ActionHalt
						}
						ui.Say(fmt.Sprintf("Deleted snapshot: %s", b.Bsu.SnapshotId))
					}
				}
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepDeregisterOMI) Cleanup(state multistep.StateBag) {
}
