package docker

import (
	"context"

	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/multistep"
)

type StepSetGeneratedData struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepSetGeneratedData) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	sha256 := "ERR_IMAGE_SHA256_NOT_FOUND"
	if imageId, ok := state.GetOk("image_id"); ok {
		s256, err := driver.Sha256(imageId.(string))
		if err == nil {
			sha256 = s256
		}
	}
	s.GeneratedData.Put("ImageSha256", sha256)
	return multistep.ActionContinue
}

func (s *StepSetGeneratedData) Cleanup(_ multistep.StateBag) {
	// No cleanup...
}
