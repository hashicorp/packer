package iso

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"fmt"
)

type FloppyConfig struct {
	FloppyIMGPath     string   `mapstructure:"floppy_img_path"`
	FloppyFiles       []string `mapstructure:"floppy_files"`
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
}

func (c *FloppyConfig) Prepare() []error {
	var errs []error

	if c.FloppyIMGPath != "" && (c.FloppyFiles != nil || c.FloppyDirectories != nil) {
		errs = append(errs,
			fmt.Errorf("'floppy_img_path' cannot be used together with 'floppy_files' and 'floppy_dirs'"),
		)
	}

	return errs
}

type StepAddFloppy struct {
	config *FloppyConfig
}

func (s *StepAddFloppy) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Adding Floppy...")

	floppyIMGPath := s.config.FloppyIMGPath
	if s.config.FloppyFiles != nil || s.config.FloppyDirectories != nil {
		var err error
		floppyIMGPath, err = s.createFloppy()
		if err != nil {
			state.Put("error", fmt.Errorf("Error creating floppy image: %v", err))
		}
	}

	vm := state.Get("vm").(*driver.VirtualMachine)
	err := vm.AddFloppy(floppyIMGPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepAddFloppy) Cleanup(state multistep.StateBag) {
	// nothing
}

func (s *StepAddFloppy) createFloppy() (string, error) {
	return "", fmt.Errorf("Not implemented")
	// TODO
}
