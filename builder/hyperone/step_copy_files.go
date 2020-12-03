package hyperone

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCopyFiles struct{}

func (s *stepCopyFiles) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if len(config.ChrootCopyFiles) == 0 {
		return multistep.ActionContinue
	}

	ui.Say("Copying files from host to chroot...")
	for _, path := range config.ChrootCopyFiles {
		chrootPath := filepath.Join(config.ChrootMountPath, path)
		log.Printf("Copying '%s' to '%s'", path, chrootPath)

		command := fmt.Sprintf("cp --remove-destination %s %s", path, chrootPath)
		err := runCommands([]string{command}, config.ctx, state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepCopyFiles) Cleanup(state multistep.StateBag) {}
