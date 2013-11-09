package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
)

// StepExport exports the container to a flat tar file.
type StepExport struct{}

func (s *StepExport) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	containerId := state.Get("container_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Exporting the container")

	// Args that we're going to pass to Docker
	args := []string{"export", containerId}

	// Open the file that we're going to write to
	f, err := os.Create(config.ExportPath)
	if err != nil {
		err := fmt.Errorf("Error creating output file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer f.Close()

	// Export the thing, take stderr and point it to the file
	var stderr bytes.Buffer
	cmd := exec.Command("docker", args...)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	log.Printf("Starting container with args: %v", args)
	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("Error exporting: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := cmd.Wait(); err != nil {
		err := fmt.Errorf("Error exporting: %s\nStderr: %s",
			err, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
