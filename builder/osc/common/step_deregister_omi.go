package common

import (
	"context"
	"fmt"
	"log"

	"github.com/antihax/optional"
	"github.com/outscale/osc-sdk-go/osc"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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

	ui := state.Get("ui").(packersdk.Ui)

	s.Regions = append(s.Regions, s.AccessConfig.GetRegion())

	log.Printf("LOG_ s.Regions: %#+v\n", s.Regions)

	for _, region := range s.Regions {
		// get new connection for each region in which we need to deregister vms
		conn := s.AccessConfig.NewOSCClientByRegion(region)

		resp, _, err := conn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
			ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
				Filters: osc.FiltersImage{
					ImageNames: []string{s.OMIName},
					//AccountAliases: []string{"self"},
				},
			}),
		})

		if err != nil {
			err := fmt.Errorf("Error describing OMI: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())

			return multistep.ActionHalt
		}

		log.Printf("LOG_ resp.Images: %#+v\n", resp.Images)

		// Deregister image(s) by name
		for i := range resp.Images {
			//We are supposing that DeleteImage does the same action as DeregisterImage
			_, _, err := conn.ImageApi.DeleteImage(context.Background(), &osc.DeleteImageOpts{
				DeleteImageRequest: optional.NewInterface(osc.DeleteImageRequest{
					ImageId: resp.Images[i].ImageId,
				}),
			})

			if err != nil {
				err := fmt.Errorf("Error deregistering existing OMI: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())

				return multistep.ActionHalt
			}

			ui.Say(fmt.Sprintf("Deregistered OMI %s, id: %s", s.OMIName, resp.Images[i].ImageId))

			// Delete snapshot(s) by image
			if s.ForceDeleteSnapshot {
				for _, b := range resp.Images[i].BlockDeviceMappings {
					if b.Bsu.SnapshotId != "" {

						_, _, err := conn.SnapshotApi.DeleteSnapshot(context.Background(), &osc.DeleteSnapshotOpts{
							DeleteSnapshotRequest: optional.NewInterface(osc.DeleteSnapshotRequest{
								SnapshotId: b.Bsu.SnapshotId,
							}),
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
