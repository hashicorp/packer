package docker

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
)

type stepBuild struct{}

func (s *stepBuild) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)

	ui := state["ui"].(packer.Ui)
	ui.Say("Building docker image '" + config.Repository + "'...")

	cmd := exec.Command("docker", "build", "-t", config.Repository, config.BuildPath)
	if err := cmd.Run(); err != nil {
		ui.Say("Failed to build image: " + config.Repository)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepBuild) Cleanup(map[string]interface{}) {}
