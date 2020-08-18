package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestOmiName     string
	ForceDeregister bool
	API             string
}

func (s *StepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.ForceDeregister {
		ui.Say("Force Deregister flag found, skipping prevalidating OMI Name")
		return multistep.ActionContinue
	}

	var (
		conn   = state.Get("osc").(*osc.APIClient)
		images []interface{}
	)

	ui.Say(fmt.Sprintf("Prevalidating OMI Name: %s", s.DestOmiName))

	resp, _, err := conn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
		ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
			Filters: osc.FiltersImage{
				ImageNames: []string{s.DestOmiName},
			},
		}),
	})

	if err != nil {
		err := fmt.Errorf("Error querying OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, omi := range resp.Images {
		if omi.ImageName == s.DestOmiName {
			images = append(images, omi)
		}
	}

	if len(images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing OMI: %s", s.DestOmiName)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, omi := range resp.Images {
		if omi.ImageName == imageName {
			images = append(images, omi)
		}
	}

	return images, nil
}

func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
