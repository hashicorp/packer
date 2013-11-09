package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"strings"
)

type StepRun struct {
	containerId string
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Starting docker container with /bin/bash")

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", "run", "-d", "-i", "-t", config.Image, "/bin/bash")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("Error running container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := cmd.Wait(); err != nil {
		err := fmt.Errorf("Error running container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.containerId = strings.TrimSpace(stdout.String())
	ui.Message(fmt.Sprintf("Container ID: %s", s.containerId))

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.containerId == "" {
		return
	}

	// TODO(mitchellh): handle errors
	exec.Command("docker", "kill", s.containerId).Run()
}
