package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"strings"
)

type StepRun struct {
	containerId string
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	tempDir := state.Get("temp_dir").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Starting docker container with /bin/bash")

	// Args that we're going to pass to Docker
	args := []string{
		"run",
		"-d", "-i", "-t",
		"-v", fmt.Sprintf("%s:/packer-files", tempDir),
		config.Image,
		"/bin/bash",
	}

	// Start the container
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Starting container with args: %v", args)
	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("Error running container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := cmd.Wait(); err != nil {
		err := fmt.Errorf("Error running container: %s\nStderr: %s",
			err, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Capture the container ID, which is alone on stdout
	s.containerId = strings.TrimSpace(stdout.String())
	ui.Message(fmt.Sprintf("Container ID: %s", s.containerId))

	state.Put("container_id", s.containerId)
	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.containerId == "" {
		return
	}

	// Kill the container. We don't handle errors because errors usually
	// just mean that the container doesn't exist anymore, which isn't a
	// big deal.
	exec.Command("docker", "kill", s.containerId).Run()
}
