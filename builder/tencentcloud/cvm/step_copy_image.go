package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCopyImage struct {
	DesinationRegions []string
	SourceRegion      string
}

func (s *stepCopyImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.DesinationRegions) == 0 || (len(s.DesinationRegions) == 1 && s.DesinationRegions[0] == s.SourceRegion) {
		return multistep.ActionContinue
	}

	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("image").(*cvm.Image).ImageId

	req := cvm.NewSyncImagesRequest()
	req.ImageIds = []*string{imageId}
	copyRegions := make([]*string, 0, len(s.DesinationRegions))
	for _, region := range s.DesinationRegions {
		if region != s.SourceRegion {
			copyRegions = append(copyRegions, &region)
		}
	}
	req.DestinationRegions = copyRegions

	_, err := client.SyncImages(req)
	if err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("copy image failed: %s", err.Error()))
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {
	// just do nothing
}
