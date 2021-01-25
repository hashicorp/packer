package yandex

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func StepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func toGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

func toBytes(gigabytesCount int) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
}

func writeSerialLogFile(ctx context.Context, state multistep.StateBag, serialLogFile string) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packersdk.Ui)

	instanceID, ok := state.GetOk("instance_id")

	if !ok || instanceID.(string) == "" {
		return nil
	}
	ui.Say("Try get instance's serial port output and write to file " + serialLogFile)
	serialOutput, err := sdk.Compute().Instance().GetSerialPortOutput(ctx, &compute.GetInstanceSerialPortOutputRequest{
		InstanceId: instanceID.(string),
	})
	if err != nil {
		return fmt.Errorf("Failed to get serial port output for instance (id: %s): %s", instanceID, err)
	}
	if err := ioutil.WriteFile(serialLogFile, []byte(serialOutput.Contents), 0600); err != nil {
		return fmt.Errorf("Failed to write serial port output to file: %s", err)
	}
	ui.Message("Serial port output has been successfully written")
	return nil
}
