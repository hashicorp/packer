package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type setpShareAlicloudImage struct {
	AlicloudImageShareAccounts   []string
	AlicloudImageUNShareAccounts []string
	RegionId                     string
}

func (s *setpShareAlicloudImage) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	for copyedRegion, copyedImageId := range alicloudImages {
		err := client.ModifyImageSharePermission(
			&ecs.ModifyImageSharePermissionArgs{
				RegionId:      common.Region(copyedRegion),
				ImageId:       copyedImageId,
				AddAccount:    s.AlicloudImageShareAccounts,
				RemoveAccount: s.AlicloudImageUNShareAccounts,
			})
		if err != nil {
			state.Put("error", err)
			ui.Say(fmt.Sprintf("Failed modifying image share permissions: %s", err))
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (s *setpShareAlicloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)
		client := state.Get("client").(*ecs.Client)
		alicloudImages := state.Get("alicloudimages").(map[string]string)
		ui.Say("Restoring image share permission because cancellations or error...")
		for copyedRegion, copyedImageId := range alicloudImages {
			err := client.ModifyImageSharePermission(
				&ecs.ModifyImageSharePermissionArgs{
					RegionId:      common.Region(copyedRegion),
					ImageId:       copyedImageId,
					AddAccount:    s.AlicloudImageUNShareAccounts,
					RemoveAccount: s.AlicloudImageShareAccounts,
				})
			if err != nil {
				ui.Say(fmt.Sprintf("Restoring image share permission failed: %s", err))
			}
		}
	}
}
