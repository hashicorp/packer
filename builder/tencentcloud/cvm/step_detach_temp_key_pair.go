package cvm

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepDetachTempKeyPair struct {
}

func (s *stepDetachTempKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)

	if _, ok := state.GetOk("temporary_key_pair_id"); !ok {
		return multistep.ActionContinue
	}

	keyId := state.Get("temporary_key_pair_id").(string)
	instance := state.Get("instance").(*cvm.Instance)

	Say(state, keyId, "Trying to detach keypair")

	req := cvm.NewDisassociateInstancesKeyPairsRequest()
	req.KeyIds = []*string{&keyId}
	req.InstanceIds = []*string{instance.InstanceId}
	req.ForceStop = common.BoolPtr(true)
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.DisassociateInstancesKeyPairs(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Fail to detach keypair from instance")
	}

	Message(state, "Waiting for keypair detached", "")
	err = WaitForInstance(ctx, client, *instance.InstanceId, "RUNNING", 1800)
	if err != nil {
		return Halt(state, err, "Failed to wait for keypair detached")
	}

	Message(state, "Keypair detached", "")

	return multistep.ActionContinue
}

func (s *stepDetachTempKeyPair) Cleanup(state multistep.StateBag) {}
