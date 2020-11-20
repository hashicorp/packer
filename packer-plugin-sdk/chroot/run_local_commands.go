package chroot

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	sl "github.com/hashicorp/packer/packer-plugin-sdk/shell-local"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

func RunLocalCommands(commands []string, wrappedCommand common.CommandWrapper, ictx interpolate.Context, ui packer.Ui) error {
	ctx := context.TODO()
	for _, rawCmd := range commands {
		intCmd, err := interpolate.Render(rawCmd, &ictx)
		if err != nil {
			return fmt.Errorf("Error interpolating: %s", err)
		}

		command, err := wrappedCommand(intCmd)
		if err != nil {
			return fmt.Errorf("Error wrapping command: %s", err)
		}

		ui.Say(fmt.Sprintf("Executing command: %s", command))
		comm := &sl.Communicator{
			ExecuteCommand: []string{"sh", "-c", command},
		}
		cmd := &packer.RemoteCmd{Command: command}
		if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
			return fmt.Errorf("Error executing command: %s", err)
		}
		if cmd.ExitStatus() != 0 {
			return fmt.Errorf(
				"Received non-zero exit code %d from command: %s",
				cmd.ExitStatus(),
				command)
		}
	}
	return nil
}
