package chroot

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepPrepareSourceImage process the source image.
type StepPrepareSourceImage struct {
	GeneratedData *packerbuilderdata.GeneratedData

	// absolute source image path
	image string
	// converted raw source image path
	rawImage string
}

func (s *StepPrepareSourceImage) prepareOutputDir(state multistep.StateBag) error {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if _, err := os.Stat(config.OutputDir); err == nil {
		if !config.PackerForce {
			return fmt.Errorf("Output directory already exists: %s", config.OutputDir)
		}

		ui.Say(fmt.Sprintf("Deleting legacy output directory \"%s\"...", config.OutputDir))
		_ = os.RemoveAll(config.OutputDir)
	}

	ui.Say(fmt.Sprintf("Creating output directory \"%s\"...", config.OutputDir))
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return err
	}

	imagePath := path.Join(config.OutputDir, config.ImageName)
	state.Put("image_path", imagePath)
	s.GeneratedData.Put("image_path", imagePath)

	return nil
}

func (s *StepPrepareSourceImage) prepareSourceImage(state multistep.StateBag) error {
	config := state.Get("config").(*Config)

	var err error

	if s.image, err = filepath.Abs(config.SourceImage); err != nil {
		return err
	}

	if _, err := os.Stat(s.image); os.IsNotExist(err) {
		return err
	}

	s.rawImage = path.Join(config.OutputDir, path.Base(s.image)) + ".raw"
	// Convert to raw format
	if _, err := RunCommand(state, fmt.Sprintf("qemu-img convert -f qcow2 -O raw  %s %s", s.image, s.rawImage)); err != nil {
		return fmt.Errorf("Cannot convert source image to raw format: %s", err)
	}

	return nil
}

func (s *StepPrepareSourceImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if err := s.prepareOutputDir(state); err != nil {
		return Halt(state, fmt.Errorf("output directory is not ready: %v", err))
	}

	if err := s.prepareSourceImage(state); err != nil {
		return Halt(state, fmt.Errorf("Source image is not ready: %v", err))
	}

	state.Put("rawImage", s.rawImage)
	s.GeneratedData.Put("rawImage", s.rawImage)

	return multistep.ActionContinue
}

func (s *StepPrepareSourceImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	if s.rawImage == "" {
		return
	}

	ui.Say("Remove raw source image...")
	_ = os.Remove(s.rawImage)

	s.rawImage = ""
}
