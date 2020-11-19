package hyperone

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	openapi "github.com/hyperonecom/h1-client-go"
)

func formatOpenAPIError(err error) string {
	openAPIError, ok := err.(openapi.GenericOpenAPIError)
	if !ok {
		return err.Error()
	}

	return fmt.Sprintf("%s (body: %s)", openAPIError.Error(), openAPIError.Body())
}

func runCommands(commands []string, ictx interpolate.Context, state multistep.StateBag) error {
	ctx := context.TODO()
	ui := state.Get("ui").(packersdk.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	comm := state.Get("communicator").(packersdk.Communicator)

	for _, rawCmd := range commands {
		intCmd, err := interpolate.Render(rawCmd, &ictx)
		if err != nil {
			return fmt.Errorf("error interpolating: %s", err)
		}

		command, err := wrappedCommand(intCmd)
		if err != nil {
			return fmt.Errorf("error wrapping command: %s", err)
		}

		remoteCmd := &packersdk.RemoteCmd{
			Command: command,
		}

		ui.Say(fmt.Sprintf("Executing command: %s", command))

		err = remoteCmd.RunWithUi(ctx, comm, ui)
		if err != nil {
			return fmt.Errorf("error running remote cmd: %s", err)
		}

		if remoteCmd.ExitStatus() != 0 {
			return fmt.Errorf(
				"received non-zero exit code %d from command: %s",
				remoteCmd.ExitStatus(),
				command)
		}
	}
	return nil
}

func captureOutput(command string, state multistep.StateBag) (string, error) {
	ctx := context.TODO()
	comm := state.Get("communicator").(packersdk.Communicator)

	var stdout bytes.Buffer
	remoteCmd := &packersdk.RemoteCmd{
		Command: command,
		Stdout:  &stdout,
	}

	log.Println(fmt.Sprintf("Executing command: %s", command))

	err := comm.Start(ctx, remoteCmd)
	if err != nil {
		return "", fmt.Errorf("error running remote cmd: %s", err)
	}

	remoteCmd.Wait()
	if remoteCmd.ExitStatus() != 0 {
		return "", fmt.Errorf(
			"received non-zero exit code %d from command: %s",
			remoteCmd.ExitStatus(),
			command)
	}

	return strings.TrimSpace(stdout.String()), nil
}
