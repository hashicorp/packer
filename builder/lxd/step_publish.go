package lxd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepPublish struct{}

func (s *stepPublish) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	name := config.ContainerName
	stop_args := []string{
		// We created the container with "--ephemeral=false" so we know it is safe to stop.
		"stop", name,
	}

	ui.Say("Stopping container...")
	_, err := LXDCommand(stop_args...)
	if err != nil {
		err := fmt.Errorf("Error stopping container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	publish_args := []string{
		"publish", name, "--alias", config.OutputImage,
	}

	for k, v := range config.PublishProperties {
		publish_args = append(publish_args, fmt.Sprintf("%s=%s", k, v))
	}

	ui.Say("Publishing container...")
	stdoutString, err := LXDCommand(publish_args...)
	if err != nil {
		err := fmt.Errorf("Error publishing container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	r := regexp.MustCompile("([0-9a-fA-F]+)$")
	fingerprint := r.FindAllStringSubmatch(stdoutString, -1)[0][0]

	ui.Say(fmt.Sprintf("Created image: %s", fingerprint))

	state.Put("imageFingerprint", fingerprint)

	return multistep.ActionContinue
}

func (s *stepPublish) Cleanup(state multistep.StateBag) {}
