package vminstance

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepAttachDataVolume struct {
	attached bool
}

func (s *StepAttachDataVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start attach data volume...")

	err := attachVolume(state)
	if err != nil {
		return halt(state, err, "")
	}
	s.attached = true

	return multistep.ActionContinue
}

func attachVolume(state multistep.StateBag) error {
	driver, _, ui := GetCommonFromState(state)

	volume := state.Get(DataVolume).(*zstacktype.DataVolume)
	vm := state.Get(Vm).(*zstacktype.VmInstance)

	deviceID, err := driver.AttachDataVolume(volume.Uuid, vm.Uuid)
	if err != nil {
		return err
	}

	volume.DeviceId = deviceID
	state.Put(DataVolume, volume)
	ui.Message(fmt.Sprintf("attach volume to vm, deviceId: %s", deviceID))

	return nil
}

func (s *StepAttachDataVolume) Cleanup(state multistep.StateBag) {
	driver, config, ui := GetCommonFromState(state)
	ui.Say("cleanup attach data volume executing...")
	if s.attached && !config.SkipDeleteVm {
		volume, ok := state.GetOk(DataVolume)
		if ok {
			driver.DetachDataVolume(volume.(*zstacktype.DataVolume).Uuid)
		}
	}
}
