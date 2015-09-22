package brkt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateImage struct {
	ImageName string
}

func timestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func (s *stepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	api := state.Get("api").(*brkt.API)
	ui := state.Get("ui").(packer.Ui)

	instance, ok := state.Get("instance").(*brkt.Instance)
	if !ok {
		state.Put("error", fmt.Errorf("error retrieving instance"))
		return multistep.ActionHalt
	}

	if _, err := instance.Reload(); err != nil {
		state.Put("error", fmt.Errorf("error reloading instance: %s", err))
		return multistep.ActionHalt
	}
	for instance.Data.ProviderInstance.State != "READY" {
		if instance.Data.ProviderInstance.State == "FAILED" {
			state.Put("error", fmt.Errorf("instance moved into FAILED state"))
			return multistep.ActionHalt
		}

		ui.Say("Waiting for instance to become ready...")
		time.Sleep(15 * time.Second)
		if _, err := instance.Reload(); err != nil {
			state.Put("error", fmt.Errorf("error reloading instance: %s", err))
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating image from provisioned workload")

	imageCreateData, err := instance.CreateImage(&brkt.InstanceCreateImagePayload{
		ImageName: s.ImageName,
	})

	if err != nil {
		state.Put("error", fmt.Errorf("error creating image: %s", err))
		return multistep.ActionHalt
	}

	imageDefinition := &brkt.ImageDefinition{
		Data: &brkt.ImageDefinitionData{
			Id: imageCreateData.RequestId,
		},
		ApiClient: api.ApiClient,
	}

	imageDefinition.Reload()

	for imageDefinition.Data.State != "READY" {
		if imageDefinition.Data.State == "FAILED" {
			state.Put("error", fmt.Errorf("Creating an image failed..."))
			return multistep.ActionHalt
		}

		ui.Say("Waiting for image definition to become ready")
		time.Sleep(15 * time.Second)
		imageDefinition.Reload()
	}

	state.Put("imageId", imageDefinition.Data.Id)
	state.Put("imageName", imageDefinition.Data.Name)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(multistep.StateBag) {}
