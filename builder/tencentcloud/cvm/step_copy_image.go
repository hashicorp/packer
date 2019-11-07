package cvm

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCopyImage struct {
	DesinationRegions []string
	SourceRegion      string
}

func (s *stepCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.DesinationRegions) == 0 || (len(s.DesinationRegions) == 1 && s.DesinationRegions[0] == s.SourceRegion) {
		return multistep.ActionContinue
	}

	client := state.Get("cvm_client").(*cvm.Client)

	imageId := state.Get("image").(*cvm.Image).ImageId

	Say(state, strings.Join(s.DesinationRegions, ","), "Trying to copy image to")

	req := cvm.NewSyncImagesRequest()
	req.ImageIds = []*string{imageId}
	copyRegions := make([]*string, 0, len(s.DesinationRegions))
	for _, region := range s.DesinationRegions {
		if region != s.SourceRegion {
			copyRegions = append(copyRegions, common.StringPtr(region))
		}
	}
	req.DestinationRegions = copyRegions

	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.SyncImages(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to copy image")
	}

	Message(state, "Image copied", "")

	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {}
