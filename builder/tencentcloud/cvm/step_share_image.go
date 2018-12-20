package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepShareImage struct {
	ShareAccounts []string
}

func (s *stepShareImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.ShareAccounts) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("image").(*cvm.Image).ImageId

	req := cvm.NewModifyImageSharePermissionRequest()
	req.ImageId = imageId
	SHARE := "SHARE"
	req.Permission = &SHARE
	accounts := make([]*string, 0, len(s.ShareAccounts))
	for _, account := range s.ShareAccounts {
		accounts = append(accounts, &account)
	}
	req.AccountIds = accounts

	_, err := client.ModifyImageSharePermission(req)
	if err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("share image failed: %s", err.Error()))
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepShareImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)
		client := state.Get("cvm_client").(*cvm.Client)
		imageId := state.Get("image").(*cvm.Image).ImageId
		ui.Say("Cancel share image due to action cancelled or halted.")

		req := cvm.NewModifyImageSharePermissionRequest()
		req.ImageId = imageId
		CANCEL := "CANCEL"
		req.Permission = &CANCEL
		accounts := make([]*string, 0, len(s.ShareAccounts))
		for _, account := range s.ShareAccounts {
			accounts = append(accounts, &account)
		}
		req.AccountIds = accounts

		_, err := client.ModifyImageSharePermission(req)
		if err != nil {
			ui.Error(fmt.Sprintf("Cancel share image failed: %s", err.Error()))
		}
	}
}
