package lxc

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepLxcCreate struct{}

func (s *stepLxcCreate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	// TODO: read from env
	lxc_dir := "/var/lib/lxc"
	rootfs := filepath.Join(lxc_dir, name, "rootfs")

	if config.PackerForce {
		s.Cleanup(state)
	}

	commands := make([][]string, 3)
	commands[0] = append(config.EnvVars, "lxc-create")
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
		log.Printf("Executing sudo command: %#v", command)
		err := s.SudoCommand(command...)
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
	ui := state.Get("ui").(packer.Ui)

	command := []string{
		"lxc-destroy", "-f", "-n", config.ContainerName,
	}

	ui.Say("Unregistering and deleting virtual machine...")
	if err := s.SudoCommand(command...); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}

func (s *stepLxcCreate) SudoCommand(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing sudo command: %#v", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Sudo command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}
