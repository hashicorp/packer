package lxc

import (
	"context"
	"fmt"
	"log"
	"os/user"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepLxcCreate struct{}

func (s *stepLxcCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	name := config.ContainerName

	// TODO: read from env
	lxc_dir := "/var/lib/lxc"
	user, err := user.Current()
	if err != nil {
		log.Print("Cannot find current user. Falling back to /var/lib/lxc...")
	}
	if user.Uid != "0" && user.HomeDir != "" {
		lxc_dir = filepath.Join(user.HomeDir, ".local", "share", "lxc")
	}
	rootfs := filepath.Join(lxc_dir, name, "rootfs")

	if config.PackerForce {
		s.Cleanup(state)
	}

	commands := make([][]string, 3)
	commands[0] = append(commands[0], "env")
	commands[0] = append(commands[0], config.EnvVars...)
	commands[0] = append(commands[0], "lxc-create")
	commands[0] = append(commands[0], config.CreateOptions...)
	commands[0] = append(commands[0], []string{"-n", name, "-t", config.Name, "--"}...)
	commands[0] = append(commands[0], config.Parameters...)
	// prevent tmp from being cleaned on boot, we put provisioning scripts there
	// todo: wait for init to finish before moving on to provisioning instead of this
	commands[1] = []string{"touch", filepath.Join(rootfs, "tmp", ".tmpfs")}
	commands[2] = append([]string{"lxc-start"}, config.StartOptions...)
	commands[2] = append(commands[2], []string{"-d", "--name", name}...)

	ui.Say("Creating container...")
	for _, command := range commands {
		err := RunCommand(command...)
		if err != nil {
			err := fmt.Errorf("Error creating container: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("mount_path", rootfs)

	return multistep.ActionContinue
}

func (s *stepLxcCreate) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	command := []string{
		"lxc-destroy", "-f", "-n", config.ContainerName,
	}

	ui.Say("Unregistering and deleting virtual machine...")
	if err := RunCommand(command...); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
