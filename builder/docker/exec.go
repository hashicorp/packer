package docker

import (
	"os/exec"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/shell-local/localexec"
)

func runAndStream(cmd *exec.Cmd, ui packersdk.Ui) error {

	args := make([]string, len(cmd.Args)-1)
	copy(args, cmd.Args[1:])

	// Scrub password from the log output.
	capturedPassword := ""
	for i, v := range args {
		if v == "-p" || v == "--password" {
			capturedPassword = args[i+1]
			break
		}
	}

	// run local command and stream output to UI.
	return localexec.RunAndStream(cmd, ui, []string{capturedPassword})
}
