package yandex

import (
	"context"
	"fmt"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func stepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func displayInstanceStatus(sdk *ycsdk.SDK, instanceID string, ui packer.Ui) {
	instance, err := sdk.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Fail to get instance data: %s", err))
	}
	ui.Message(fmt.Sprintf("Current instance status %s", instance.Status))
}

func toGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

func toBytes(gigabytesCount int) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
}
