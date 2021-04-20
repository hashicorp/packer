package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type stepPreValidate struct {
	Force        bool
	ImageName    string
	SnapshotName string
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.Force {
		ui.Say("Force flag found, skipping prevalidating image name")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Prevalidating image name: %s", s.ImageName))

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	images, err := instanceAPI.ListImages(
		&instance.ListImagesRequest{Name: &s.ImageName},
		scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("Error: getting image list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, im := range images.Images {
		if im.Name == s.ImageName {
			err := fmt.Errorf("Error: image name: '%s' is used by existing image with ID %s",
				s.ImageName, im.ID)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say(fmt.Sprintf("Prevalidating snapshot name: %s", s.SnapshotName))

	snapshots, err := instanceAPI.ListSnapshots(
		&instance.ListSnapshotsRequest{Name: &s.SnapshotName},
		scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("Error: getting snapshot list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, sn := range snapshots.Snapshots {
		if sn.Name == s.SnapshotName {
			err := fmt.Errorf("Error: snapshot name: '%s' is used by existing snapshot with ID %s",
				s.SnapshotName, sn.ID)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}
