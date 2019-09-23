package vminstance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateVmInstance struct {
	vm *zstacktype.VmInstance
}

func (s *StepCreateVmInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start create zstack vminstance...")

	var vm *zstacktype.VmInstance
	var err error

	if vm, err = createVmInstance(state); err != nil {
		return halt(state, err, "createVmInstance")
	}
	s.vm = vm
	state.Put(Vm, s.vm)
	ui.Message(fmt.Sprintf("created vm[%s] with public ip[%s]", vm.Uuid, vm.PublicIp))

	return multistep.ActionContinue
}

func createVmInstance(state multistep.StateBag) (*zstacktype.VmInstance, error) {
	driver, config, _ := GetCommonFromState(state)

	params := &zstacktype.CreateVm{
		Name:             config.InstanceName,
		L3:               config.L3Network,
		InstanceOffering: config.InstanceOffering,
		Image:            config.Image,
		Sshkey:           string(config.Comm.SSHPublicKey),
		UserData:         config.UserData,
		Timeout:          config.stateTimeout,
	}

	vm, err := driver.CreateVmInstance(*params)
	if err != nil {
		return nil, err
	} else if vm == nil {
		return nil, fmt.Errorf("cannot find new created vm")
	}
	errCh := driver.WaitForInstance("Running", vm.Uuid, "vm")
	select {
	case err = <-errCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(config.stateTimeout):
		return nil, fmt.Errorf("time out after %v ms while waiting for vm instance to become Running", config.stateTimeout.Nanoseconds()/1000/1000)
	}
	return vm, nil
}

func (s *StepCreateVmInstance) Cleanup(state multistep.StateBag) {
	driver, config, ui := GetCommonFromState(state)
	ui.Say("cleanup create vm instance executing...")
	if s.vm != nil && !config.SkipDeleteVm {
		err := driver.DeleteVmInstance(s.vm.Uuid)
		if err != nil {
			ui.Error(err.Error())
		}
	} else {
		ui.Message("skip clean up vm instance after work")
	}
}
