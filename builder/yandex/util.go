package yandex

import (
	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func stepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
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
