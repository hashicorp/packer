package lxd

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type stepExport struct{}

func (s *stepExport) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	//outputPath := filepath.Join(config.OutputDir, "rootfs.tar.gz")

	commands := make([][]string, 3)
	commands[0] = []string{
		"lxc", "stop", name,
	}
	commands[1] = []string{
		"lxc", "export",
	}
	//commands[1] = []string{
	//	"tar", "-C", containerDir, "--numeric-owner", "--anchored", "--exclude=./rootfs/dev/log", "-czf", outputPath, "./rootfs",
	//}
	commands[2] = []string{
		"sh", "-c", "chown $USER:`id -gn` " + filepath.Join(config.OutputDir, "*"),
	}

	ui.Say("Exporting container...")
	for _, command := range commands {
		err := s.SudoCommand(command...)
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

func (s *stepExport) SudoCommand(args ...string) error {
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
