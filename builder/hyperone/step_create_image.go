package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepCreateImage struct {
	imageID string
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	vmID := state.Get("vm_id").(string)

	ui.Say("Creating image...")

	image, _, err := client.ImageApi.ImageCreate(ctx, openapi.ImageCreate{
		Name:        config.ImageName,
		Vm:          vmID,
		Service:     config.ImageService,
		Description: config.ImageDescription,
		Tag:         config.ImageTags,
	})
	if err != nil {
		err := fmt.Errorf("error creating image: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.imageID = image.Id

	state.Put("image_id", image.Id)
	state.Put("image_name", image.Name)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageID == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)

	_, err := client.ImageApi.ImageDelete(context.TODO(), s.imageID)
	if err != nil {
		ui.Error(fmt.Sprintf("error deleting image '%s' - consider deleting it manually: %s",
			s.imageID, formatOpenAPIError(err)))
	}
}
