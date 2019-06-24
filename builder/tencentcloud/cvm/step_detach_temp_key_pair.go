package cvm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepDetachTempKeyPair struct {
}

func (s *stepDetachTempKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)
	instance := state.Get("instance").(*cvm.Instance)
	if _, ok := state.GetOk("temporary_key_pair_id"); !ok {
		return multistep.ActionContinue
	}
	keyId := state.Get("temporary_key_pair_id").(*string)
	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Detaching temporary key pair %s...", *keyId))
	req := cvm.NewDisassociateInstancesKeyPairsRequest()
	req.KeyIds = []*string{keyId}
	req.InstanceIds = []*string{instance.InstanceId}
	req.ForceStop = common.BoolPtr(true)
	err := retry.Config{
		Tries: 60,
		RetryDelay: (&retry.Backoff{
			InitialBackoff: 5 * time.Second,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2,
		}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := client.DisassociateInstancesKeyPairs(req)
		return err
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Fail to detach temporary key pair from instance! Error: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepDetachTempKeyPair) Cleanup(state multistep.StateBag) {
}
