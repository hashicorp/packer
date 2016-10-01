package chroot

import (
	"fmt"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/shell-local"
	"github.com/mitchellh/packer/template/interpolate"
)

func RunLocalCommands(commands []string, wrappedCommand CommandWrapper, ctx interpolate.Context, ui packer.Ui) error {
	for _, rawCmd := range commands {
		intCmd, err := interpolate.Render(rawCmd, &ctx)
		if err != nil {
			return fmt.Errorf("Error interpolating: %s", err)
		}

		command, err := wrappedCommand(intCmd)
		if err != nil {
			return fmt.Errorf("Error wrapping command: %s", err)
		}

		ui.Say(fmt.Sprintf("Executing command: %s", command))
		comm := &shell_local.Communicator{}
		cmd := &packer.RemoteCmd{Command: command}
		if err := cmd.StartWithUi(comm, ui); err != nil {
			return fmt.Errorf("Error executing command: %s", err)
		}
		if cmd.ExitStatus != 0 {
			return fmt.Errorf(
				"Received non-zero exit code %d from command: %s",
				cmd.ExitStatus,
				command)
		}
	}
	return nil
}
