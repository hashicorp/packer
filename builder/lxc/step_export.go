package lxc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepExport struct{}

func (s *stepExport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	containerDir := fmt.Sprintf("/var/lib/lxc/%s", name)
	outputPath := filepath.Join(config.OutputDir, "rootfs.tar.gz")
	configFilePath := filepath.Join(config.OutputDir, "lxc-config")

	configFile, err := os.Create(configFilePath)

	if err != nil {
		err := fmt.Errorf("Error creating config file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	originalConfigFile, err := os.Open(config.ConfigFile)

	if err != nil {
		err := fmt.Errorf("Error opening config file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, err = io.Copy(configFile, originalConfigFile)

	commands := make([][]string, 3)
	commands[0] = []string{
		"lxc-stop", "--name", name,
	}
	commands[1] = []string{
		"tar", "-C", containerDir, "--numeric-owner", "--anchored", "--exclude=./rootfs/dev/log", "-czf", outputPath, "./rootfs",
	}
	commands[2] = []string{
		"chmod", "+x", configFilePath,
	}

	ui.Say("Exporting container...")
	for _, command := range commands {
		err := RunCommand(command...)
		if err != nil {
			err := fmt.Errorf("Error exporting container: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepExport) Cleanup(state multistep.StateBag) {}
