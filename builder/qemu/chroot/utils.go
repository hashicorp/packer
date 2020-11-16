package chroot

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func Halt(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func RunCommand(state multistep.StateBag, cmd string) (string, error) {
	cmd, err := state.Get("wrappedCommand").(common.CommandWrapper)(cmd)

	if err != nil {
		return "", err
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Running command \"%s\"...", cmd))

	shell := common.ShellCommand(cmd)
	shell.Stderr = new(bytes.Buffer)
	shell.Stdout = new(bytes.Buffer)

	if err := shell.Run(); err != nil {
		return fmt.Sprintf("%v", shell.Stdout), fmt.Errorf("%v", shell.Stderr)
	}

	return strings.TrimSpace(fmt.Sprintf("%v", shell.Stdout)), nil
}
