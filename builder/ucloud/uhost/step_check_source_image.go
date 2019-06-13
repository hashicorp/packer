package uhost

import (
	"context"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCheckSourceImageId struct {
	SourceUHostImageId string
}

func (s *stepCheckSourceImageId) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*UCloudClient)

	ui.Say("Querying source image id...")

	imageSet, err := client.DescribeImageById(s.SourceUHostImageId)
	if err != nil {
		if isNotFoundError(err) {
			return halt(state, err, "")
		}
		return halt(state, err, "Error on querying source_image_id")
	}

	if imageSet.OsType == osTypeWindows {
		return halt(state, err, "The builder of ucloud-uhost not support build Windows image yet")
	}

	state.Put("source_image", imageSet)
	return multistep.ActionContinue
}

func (s *stepCheckSourceImageId) Cleanup(multistep.StateBag) {}
