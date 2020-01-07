package uhost

import (
	"context"
	"fmt"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCheckSourceImageId struct {
	SourceUHostImageId string
}

func (s *stepCheckSourceImageId) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*ucloudcommon.UCloudClient)

	ui.Say("Querying source image id...")

	imageSet, err := client.DescribeImageById(s.SourceUHostImageId)
	if err != nil {
		if ucloudcommon.IsNotFoundError(err) {
			return ucloudcommon.Halt(state, err, "")
		}
		return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on querying specified source_image_id %q", s.SourceUHostImageId))
	}

	if imageSet.OsType == ucloudcommon.OsTypeWindows {
		return ucloudcommon.Halt(state, err, "The ucloud-uhost builder does not support Windows images yet")
	}

	state.Put("source_image", imageSet)
	return multistep.ActionContinue
}

func (s *stepCheckSourceImageId) Cleanup(multistep.StateBag) {}
