package yandex

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type stepCreateImage struct{}

func (stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	diskID := state.Get("disk_id").(string)

	ui.Say(fmt.Sprintf("Creating image: %v", c.ImageName))
	ctx, cancel := context.WithTimeout(ctx, c.StateTimeout)
	defer cancel()

	op, err := sdk.WrapOperation(sdk.Compute().Image().Create(ctx, &compute.CreateImageRequest{
		FolderId:    c.FolderID,
		Name:        c.ImageName,
		Family:      c.ImageFamily,
		Description: c.ImageDescription,
		Labels:      c.ImageLabels,
		ProductIds:  c.ImageProductIDs,
		Source: &compute.CreateImageRequest_DiskId{
			DiskId: diskID,
		},
	}))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error creating image: %s", err))
	}

	ui.Say("Waiting for image to complete...")
	if err := op.Wait(ctx); err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error waiting for image: %s", err))
	}

	resp, err := op.Response()
	if err != nil {
		return stepHaltWithError(state, err)
	}

	image, ok := resp.(*compute.Image)
	if !ok {
		return stepHaltWithError(state, errors.New("Response doesn't contain Image"))
	}

	log.Printf("Image ID: %s", image.Id)
	log.Printf("Image Name: %s", image.Name)
	log.Printf("Image Family: %s", image.Family)
	log.Printf("Image Description: %s", image.Description)
	log.Printf("Image Storage size: %d", image.StorageSize)
	state.Put("image", image)

	return multistep.ActionContinue
}

func (stepCreateImage) Cleanup(state multistep.StateBag) {
	// no cleanup
}
