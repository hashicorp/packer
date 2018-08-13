package lxc

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepExport struct{}

func (s *stepExport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	lxc_dir := "/var/lib/lxc"
	user, err := user.Current()
	if err != nil {
		log.Print("Cannot find current user. Falling back to /var/lib/lxc...")
	}
	if user.Uid != "0" && user.HomeDir != "" {
		lxc_dir = filepath.Join(user.HomeDir, ".local", "share", "lxc")
	}

	containerDir := filepath.Join(lxc_dir, name)
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
