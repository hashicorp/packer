package vminstance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateImage struct {
	images []*zstacktype.Image
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start create zstack image...")

	images, err := createImage(state)
	if err != nil {
		return halt(state, err, "")
	}
	s.images = images
	state.Put(Image, s.images)

	return multistep.ActionContinue
}

func createImage(state multistep.StateBag) ([]*zstacktype.Image, error) {
	driver, config, ui := GetCommonFromState(state)

	v := state.Get(DataVolume)
	var err error
	images := []*zstacktype.Image{}
	var image *zstacktype.Image
	if v != nil {
		dataVolumeUuid := (v.(*zstacktype.DataVolume)).Uuid
		params := &zstacktype.CreateVolumeImage{
			Name:          config.ImageName + "-Data",
			DataVolume:    dataVolumeUuid,
			BackupStorage: state.Get(BackupStorage).(string),
			Timeout:       config.stateTimeout,
		}
		image, err = driver.CreateDataVolumeImage(*params)
		errCh := driver.WaitForInstance("Ready", image.Uuid, "image")
		select {
		case err = <-errCh:
			if err != nil {
				return nil, err
			}
		case <-time.After(config.stateTimeout):
			return nil, fmt.Errorf("time out after %v ms while waiting for image to become Ready", config.stateTimeout.Nanoseconds()/1000/1000)
		}
		images = append(images, image)
	}
	if v == nil || config.CreateWithRoot {
		params := &zstacktype.CreateImage{
			Name:          config.ImageName + "-Root",
			GusetOsType:   state.Get(GuestOSType).(string),
			Platform:      state.Get(Platform).(string),
			RootVolume:    (state.Get(Vm).(*zstacktype.VmInstance)).RootVolume,
			BackupStorage: state.Get(BackupStorage).(string),
			Timeout:       config.stateTimeout,
		}
		image, err = driver.CreateImage(*params)
		errCh := driver.WaitForInstance("Ready", image.Uuid, "image")
		select {
		case err = <-errCh:
			if err != nil {
				return nil, err
			}
		case <-time.After(config.stateTimeout):
			return nil, fmt.Errorf("time out after %v ms while waiting for image to become Ready", config.stateTimeout.Nanoseconds()/1000/1000)
		}
		images = append(images, image)
	}

	if err != nil {
		return nil, err
	} else if len(images) == 0 {
		return nil, fmt.Errorf("cannot find new created image")
	}

	for _, v := range images {
		ui.Message(fmt.Sprintf("created image: uuid[%s], name[%s]", v.Uuid, v.Name))
	}

	return images, nil
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup create image executing...")
}
