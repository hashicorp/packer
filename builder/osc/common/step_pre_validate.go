package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestOmiName     string
	ForceDeregister bool
}

func (s *StepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.ForceDeregister {
		ui.Say("Force Deregister flag found, skipping prevalidating OMI Name")
		return multistep.ActionContinue
	}

	oapiconn := state.Get("oapi").(*oapi.Client)

	ui.Say(fmt.Sprintf("Prevalidating OMI Name: %s", s.DestOmiName))
	resp, err := oapiconn.POST_ReadImages(oapi.ReadImagesRequest{
		Filters: oapi.FiltersImage{ImageNames: []string{s.DestOmiName}},
	})

	if err != nil {
		err := fmt.Errorf("Error querying OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//FIXME: Remove when the oAPI filters works
	images := make([]oapi.Image, 0)

	for _, omi := range resp.OK.Images {
		if omi.ImageName == s.DestOmiName {
			images = append(images, omi)
		}
	}

	//if len(resp.OK.Images) > 0 {
	if len(images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing OMI: %s", resp.OK.Images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
