package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

type StepUpdateOMIAttributes struct {
	AccountIds         []string
	SnapshotAccountIds []string
	Ctx                interpolate.Context
}

func (s *StepUpdateOMIAttributes) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	config := state.Get("clientConfig").(*oapi.Config)
	ui := state.Get("ui").(packer.Ui)
	omis := state.Get("omis").(map[string]string)
	snapshots := state.Get("snapshots").(map[string][]string)

	// Determine if there is any work to do.
	valid := false
	valid = valid || (s.AccountIds != nil && len(s.AccountIds) > 0)
	valid = valid || (s.SnapshotAccountIds != nil && len(s.SnapshotAccountIds) > 0)

	if !valid {
		return multistep.ActionContinue
	}

	s.Ctx.Data = extractBuildInfo(oapiconn.GetConfig().Region, state)

	updateSnapshoptRequest := oapi.UpdateSnapshotRequest{
		PermissionsToCreateVolume: oapi.PermissionsOnResourceCreation{
			Additions: oapi.PermissionsOnResource{
				AccountIds:       s.AccountIds,
				GlobalPermission: false,
			},
		},
	}

	updateImageRequest := oapi.UpdateImageRequest{
		PermissionsToLaunch: oapi.PermissionsOnResourceCreation{
			Additions: oapi.PermissionsOnResource{
				AccountIds:       s.AccountIds,
				GlobalPermission: false,
			},
		},
	}

	// Updating image attributes
	for region, omi := range omis {
		ui.Say(fmt.Sprintf("Updating attributes on OMI (%s)...", omi))
		newConfig := &oapi.Config{
			UserAgent: config.UserAgent,
			AccessKey: config.AccessKey,
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

		ui.Message(fmt.Sprintf("Updating: %s", omi))
		updateImageRequest.ImageId = omi
		_, err := regionconn.POST_UpdateImage(updateImageRequest)
		if err != nil {
			err := fmt.Errorf("Error updating OMI: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Updating snapshot attributes
	for region, region_snapshots := range snapshots {
		for _, snapshot := range region_snapshots {
			ui.Say(fmt.Sprintf("Updating attributes on snapshot (%s)...", snapshot))
			newConfig := &oapi.Config{
				UserAgent: config.UserAgent,
				AccessKey: config.AccessKey,
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

			ui.Message(fmt.Sprintf("Updating: %s", snapshot))
			updateSnapshoptRequest.SnapshotId = snapshot
			_, err := regionconn.POST_UpdateSnapshot(updateSnapshoptRequest)
			if err != nil {
				err := fmt.Errorf("Error updating snapshot: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		}
	}

	return multistep.ActionContinue
}

func (s *StepUpdateOMIAttributes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
