package cvm

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepShareImage struct {
	ShareAccounts []string
}

func (s *stepShareImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.ShareAccounts) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("cvm_client").(*cvm.Client)

	imageId := state.Get("image").(*cvm.Image).ImageId
	Say(state, strings.Join(s.ShareAccounts, ","), "Trying to share image to")

	req := cvm.NewModifyImageSharePermissionRequest()
	req.ImageId = imageId
	req.Permission = common.StringPtr("SHARE")
	accounts := make([]*string, 0, len(s.ShareAccounts))
	for _, account := range s.ShareAccounts {
		accounts = append(accounts, common.StringPtr(account))
	}
	req.AccountIds = accounts
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.ModifyImageSharePermission(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to share image")
	}

	Message(state, "Image shared", "")

	return multistep.ActionContinue
}

func (s *stepShareImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ctx := context.TODO()
	client := state.Get("cvm_client").(*cvm.Client)

	imageId := state.Get("image").(*cvm.Image).ImageId
	SayClean(state, "image share")

	req := cvm.NewModifyImageSharePermissionRequest()
	req.ImageId = imageId
	req.Permission = common.StringPtr("CANCEL")
	accounts := make([]*string, 0, len(s.ShareAccounts))
	for _, account := range s.ShareAccounts {
		accounts = append(accounts, &account)
	}
	req.AccountIds = accounts
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.ModifyImageSharePermission(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to cancel share image(%s), please delete it manually", *imageId))
	}
}
