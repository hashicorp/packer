package lxd

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"strings"
	"time"
)

type stepLxdLaunch struct{}

func (s *stepLxdLaunch) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName
	image := config.Image

	commands := [][]string{
		{"lxc", "launch", image, name},
	}

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
	// TODO: Should we check `lxc info <container>` for "Running"?
	// We have to do this so /tmp doens't get cleared and lose our provisioner scripts.
	time.Sleep(1 * time.Second)

	return multistep.ActionContinue
}

func (s *stepLxdLaunch) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	command := []string{
		"lxc", "delete", "--force", config.ContainerName,
	}

	ui.Say("Unregistering and deleting deleting container...")
	if err := s.SudoCommand(command...); err != nil {
		ui.Error(fmt.Sprintf("Error deleting container: %s", err))
	}
}

func (s *stepLxdLaunch) SudoCommand(args ...string) error {
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
