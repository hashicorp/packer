package docker

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

// StepExport exports the container to a flat tar file.
type StepExport struct{}

func (s *StepExport) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	containerId := state.Get("container_id").(string)
	ui := state.Get("ui").(packer.Ui)

	// Open the file that we're going to write to
	f, err := os.Create(config.ExportPath)
	if err != nil {
		err := fmt.Errorf("Error creating output file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Exporting the container")
	if err := driver.Export(containerId, f); err != nil {
		f.Close()
		os.Remove(f.Name())

		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	f.Close()
	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
