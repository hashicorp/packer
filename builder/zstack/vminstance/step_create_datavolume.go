package vminstance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepCreateDataVolume struct {
	volume *zstacktype.DataVolume
}

func (s *StepCreateDataVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start create data volume...")

	var volume *zstacktype.DataVolume
	var err error

	if volume, err = createVolume(state); err != nil {
		return halt(state, err, "")
	}
	s.volume = volume
	state.Put(DataVolume, s.volume)

	return multistep.ActionContinue
}

func createVolume(state multistep.StateBag) (*zstacktype.DataVolume, error) {
	driver, config, ui := GetCommonFromState(state)

	vm := state.Get(Vm).(*zstacktype.VmInstance)
	root, err := driver.QueryVolume(vm.RootVolume)
	if err != nil {
		return nil, err
	}

	name, err := interpolate.Render("packer-{{timestamp}}", nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse volume name: %s ", err)
	}

	var volume *zstacktype.DataVolume

	if config.DataVolumeSize != "" {
		size, err1 := getSizeFromStr(config.DataVolumeSize)
		if err1 != nil {
			return nil, err1
		}
		params := &zstacktype.CreateDataVolume{
			Size:           size,
			PrimaryStorage: root.PrimaryStorage,
			Name:           name,
			Host:           vm.Host,
			Timeout:        config.stateTimeout,
		}
		volume, err = driver.CreateDataVolumeFromSize(*params)
	} else if config.DataVolumeImage != "" {
		params := &zstacktype.CreateDataVolumeFromImage{
			Uuid:           config.DataVolumeImage,
			PrimaryStorage: root.PrimaryStorage,
			Name:           name,
			Host:           vm.Host,
			Timeout:        config.stateTimeout,
		}
		volume, err = driver.CreateDataVolumeFromImage(*params)
	} else {
		return nil, fmt.Errorf("can not be here!")
	}

	if err != nil {
		return nil, err
	} else if volume == nil {
		return nil, fmt.Errorf("cannot find new created volume")
	}

	errCh := driver.WaitForInstance("Ready", volume.Uuid, "volume")
	select {
	case err = <-errCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(config.stateTimeout):
		return nil, fmt.Errorf("time out after %v ms while waiting for volume to become Ready", config.stateTimeout.Nanoseconds()/1000/1000)
	}
	ui.Message(fmt.Sprintf("created volume: uuid[%s], name[%s]", volume.Uuid, volume.Name))

	return volume, nil
}

func (s *StepCreateDataVolume) Cleanup(state multistep.StateBag) {
	driver, config, ui := GetCommonFromState(state)
	ui.Say("cleanup create data volume executing...")
	if s.volume != nil && !config.SkipDeleteVm {
		err := driver.DeleteDataVolume(s.volume.Uuid)
		if err != nil {
			ui.Error(err.Error())
		}
	} else {
		ui.Message("skip clean up data volume after work")
	}
}
