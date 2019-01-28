package hyperone

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/hyperonecom/h1-client-go"
)

func formatOpenAPIError(err error) string {
	openAPIError, ok := err.(openapi.GenericOpenAPIError)
	if !ok {
		return err.Error()
	}

	return fmt.Sprintf("%s (body: %s)", openAPIError.Error(), openAPIError.Body())
}

func runCommands(commands []string, ctx interpolate.Context, state multistep.StateBag) error {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	comm := state.Get("communicator").(packer.Communicator)

	for _, rawCmd := range commands {
		intCmd, err := interpolate.Render(rawCmd, &ctx)
		if err != nil {
			return fmt.Errorf("error interpolating: %s", err)
		}

		command, err := wrappedCommand(intCmd)
		if err != nil {
			return fmt.Errorf("error wrapping command: %s", err)
		}

		remoteCmd := &packer.RemoteCmd{
			Command: command,
		}

		ui.Say(fmt.Sprintf("Executing command: %s", command))

		err = remoteCmd.StartWithUi(comm, ui)
		if err != nil {
			return fmt.Errorf("error running remote cmd: %s", err)
		}

		if remoteCmd.ExitStatus != 0 {
			return fmt.Errorf(
				"received non-zero exit code %d from command: %s",
				remoteCmd.ExitStatus,
				command)
		}
	}
	return nil
}
