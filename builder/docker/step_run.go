package docker

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRun struct {
	containerId string
}

func (s *StepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config, ok := state.Get("config").(*Config)
	if !ok {
		err := fmt.Errorf("error encountered obtaining docker config")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	runConfig := ContainerConfig{
		Image:      config.Image,
		RunCommand: config.RunCommand,
		Device:     config.Device,
		TmpFs:      config.TmpFs,
		Volumes:    make(map[string]string),
		CapAdd:     config.CapAdd,
		CapDrop:    config.CapDrop,
		Privileged: config.Privileged,
	}

	for host, container := range config.Volumes {
		runConfig.Volumes[host] = container
	}

	tempDir := state.Get("temp_dir").(string)
	runConfig.Volumes[tempDir] = config.ContainerDir

	driver := state.Get("driver").(Driver)
	ui.Say("Starting docker container...")
	containerId, err := driver.StartContainer(&runConfig)
	if err != nil {
		err := fmt.Errorf("Error running container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Save the container ID
	s.containerId = containerId
	state.Put("container_id", s.containerId)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.containerId)
	ui.Message(fmt.Sprintf("Container ID: %s", s.containerId))
	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.containerId == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// Kill the container. We don't handle errors because errors usually
	// just mean that the container doesn't exist anymore, which isn't a
	// big deal.
	ui.Say(fmt.Sprintf("Killing the container: %s", s.containerId))
	driver.KillContainer(s.containerId)

	// Reset the container ID so that we're idempotent
	s.containerId = ""
}
