package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCreateImage struct {
	imageId string
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)

	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*cvm.Instance)

	Say(state, config.ImageName, "Trying to create a new image")

	req := cvm.NewCreateImageRequest()
	req.ImageName = &config.ImageName
	req.ImageDescription = &config.ImageDescription
	req.InstanceId = instance.InstanceId

	// TODO: We should allow user to specify which data disk should be
	// included into created image.
	var dataDiskIds []*string
	for _, disk := range instance.DataDisks {
		dataDiskIds = append(dataDiskIds, disk.DiskId)
	}
	if len(dataDiskIds) > 0 {
		req.DataDiskIds = dataDiskIds
	}

	True := "True"
	False := "False"
	if config.ForcePoweroff {
		req.ForcePoweroff = &True
	} else {
		req.ForcePoweroff = &False
	}

	if config.Sysprep {
		req.Sysprep = &True
	} else {
		req.Sysprep = &False
	}

	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.CreateImage(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create image")
	}

	Message(state, "Waiting for image ready", "")
	err = WaitForImageReady(ctx, client, config.ImageName, "NORMAL", 3600)
	if err != nil {
		return Halt(state, err, "Failed to wait for image ready")
	}

	image, err := GetImageByName(ctx, client, config.ImageName)
	if err != nil {
		return Halt(state, err, "Failed to get image")
	}

	if image == nil {
		return Halt(state, fmt.Errorf("No image return"), "Failed to crate image")
	}

	s.imageId = *image.ImageId
	state.Put("image", image)
	Message(state, s.imageId, "Image created")

	tencentCloudImages := make(map[string]string)
	tencentCloudImages[config.Region] = s.imageId
	state.Put("tencentcloudimages", tencentCloudImages)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageId == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ctx := context.TODO()
	client := state.Get("cvm_client").(*cvm.Client)

	SayClean(state, "image")

	req := cvm.NewDeleteImagesRequest()
	req.ImageIds = []*string{&s.imageId}
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.DeleteImages(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to delete image(%s), please delete it manually", s.imageId))
	}
}
