package wim

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	installWIM     = `install.wim`
	installWIMPath = `sources\` + installWIM
)

type StepExtractWIM struct {
	DevicePathKey string
	SkipOperation bool
	WIMPathKey    string
}

func (s *StepExtractWIM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipOperation {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	buildDir := state.Get("build_dir").(string)
	devicePath := state.Get(s.DevicePathKey).(string)

	ui.Say("Extracting WIM...")

	// Copy WIM to the temp directory

	srcWIMPath := fmt.Sprintf("%s\\%s", devicePath, installWIMPath)

	_, err := os.Stat(srcWIMPath)
	if os.IsNotExist(err) {
		err = fmt.Errorf("Error gathering informabout about WIM: %s", srcWIMPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	srcWIM, err := os.Open(srcWIMPath)
	if err != nil {
		err = fmt.Errorf("Error opening source WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer srcWIM.Close()

	dstWIMPath := fmt.Sprintf("%s/%s", buildDir, installWIM)

	dstWIM, err := os.Create(dstWIMPath)
	if err != nil {
		err = fmt.Errorf("Error opening destination WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer dstWIM.Close()

	if _, err = io.Copy(dstWIM, srcWIM); err != nil {
		err = fmt.Errorf("Error copying WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Update state bag
	state.Put(s.WIMPathKey, dstWIMPath)

	return multistep.ActionContinue
}

func (s *StepExtractWIM) Cleanup(state multistep.StateBag) {
	if s.SkipOperation {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	// Remove copied WIM
	if wimPath, ok := state.GetOk(s.WIMPathKey); ok {
		if err := os.Remove(wimPath.(string)); err != nil {
			err = fmt.Errorf("Error deleting WIM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
		} else {
			ui.Say(fmt.Sprintf("Removed WIM %s", wimPath))

			// Update state bag
			state.Remove(s.WIMPathKey)
		}
	}
}
