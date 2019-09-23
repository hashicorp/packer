package vminstance

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepExportImage struct {
}

func (s *StepExportImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("start export zstack image...")

	paths, err := exportImage(state)
	if err != nil {
		return halt(state, err, "")
	}

	state.Put(ExportPath, paths)
	for _, path := range paths {
		ui.Message(fmt.Sprintf("export image to: %s", path))
	}

	return multistep.ActionContinue
}

func exportImage(state multistep.StateBag) ([]string, error) {
	driver, _, _ := GetCommonFromState(state)

	images := state.Get(Image).([]*zstacktype.Image)
	paths := []string{}
	for _, image := range images {
		path, err := driver.ExportImage(*image)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func (s *StepExportImage) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup export image executing...")
}
