package hyperone

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	openapi "github.com/hyperonecom/h1-client-go"
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

func captureOutput(command string, state multistep.StateBag) (string, error) {
	comm := state.Get("communicator").(packer.Communicator)

	var stdout bytes.Buffer
	remoteCmd := &packer.RemoteCmd{
		Command: command,
		Stdout:  &stdout,
	}

	log.Println(fmt.Sprintf("Executing command: %s", command))

	err := comm.Start(remoteCmd)
	if err != nil {
		return "", fmt.Errorf("error running remote cmd: %s", err)
	}

	remoteCmd.Wait()
	if remoteCmd.ExitStatus != 0 {
		return "", fmt.Errorf(
			"received non-zero exit code %d from command: %s",
			remoteCmd.ExitStatus,
			command)
	}

	return strings.TrimSpace(stdout.String()), nil
}
