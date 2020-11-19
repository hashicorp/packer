package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepShareAlicloudImage struct {
	AlicloudImageShareAccounts   []string
	AlicloudImageUNShareAccounts []string
	RegionId                     string
}

func (s *stepShareAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	alicloudImages := state.Get("alicloudimages").(map[string]string)

	for regionId, imageId := range alicloudImages {
		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.AddAccount = &s.AlicloudImageShareAccounts
		modifyImageShareRequest.RemoveAccount = &s.AlicloudImageUNShareAccounts

		if _, err := client.ModifyImageSharePermission(modifyImageShareRequest); err != nil {
			return halt(state, err, "Failed modifying image share permissions")
		}
	}
	return multistep.ActionContinue
}

func (s *stepShareAlicloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*ClientWrapper)
	alicloudImages := state.Get("alicloudimages").(map[string]string)

	ui.Say("Restoring image share permission because cancellations or error...")

	for regionId, imageId := range alicloudImages {
		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.AddAccount = &s.AlicloudImageUNShareAccounts
		modifyImageShareRequest.RemoveAccount = &s.AlicloudImageShareAccounts
		if _, err := client.ModifyImageSharePermission(modifyImageShareRequest); err != nil {
			ui.Say(fmt.Sprintf("Restoring image share permission failed: %s", err))
		}
	}
}
