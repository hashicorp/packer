package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepExport exports the container to a flat tar file.
type StepExport struct{}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config, ok := state.Get("config").(*Config)
	if !ok {
		err := fmt.Errorf("error encountered obtaining docker config")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We should catch this in validation, but guard anyway
	if config.ExportPath == "" {
		err := fmt.Errorf("No output file specified, we can't export anything")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Make the directory we're exporting to if it doesn't exist
	exportDir := filepath.Dir(config.ExportPath)
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Open the file that we're going to write to
	f, err := os.Create(config.ExportPath)
	if err != nil {
		err := fmt.Errorf("Error creating output file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	driver := state.Get("driver").(Driver)
	containerId := state.Get("container_id").(string)

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
