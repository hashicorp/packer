// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
)

const DefaultRemotePath = "/tmp/script.sh"

// TODO(mitchellh): config
type config struct {
	// The local path of the shell script to upload and execute.
	Path string

	// The remote path where the local shell script will be uploaded to.
	// This should be set to a writable file that is in a pre-existing directory.
	RemotePath string
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raw interface{}, ui packer.Ui) {
	// TODO: errors
	_ = mapstructure.Decode(raw, &p.config)

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	ui.Say(fmt.Sprintf("Provisioning with shell script: %s", p.config.Path))

	log.Printf("Opening %s for reading", p.config.Path)
	f, err := os.Open(p.config.Path)
	if err != nil {
		ui.Error(fmt.Sprintf("Error opening shell script: %s", err))
		return
	}

	log.Printf("Uploading %s => %s", p.config.Path, p.config.RemotePath)
	err = comm.Upload(p.config.RemotePath, f)
	if err != nil {
		ui.Error(fmt.Sprintf("Error uploading shell script: %s", err))
		return
	}

	command := fmt.Sprintf("chmod +x %s && %s", p.config.RemotePath, p.config.RemotePath)
	log.Printf("Executing command: %s", command)
	cmd, err := comm.Start(command)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed executing command: %s", err))
		return
	}

	exit := cmd.ExitChan()
	stderr := cmd.StderrChan()
	stdout := cmd.StdoutChan()

	for {
		select {
		case output := <-stderr:
			ui.Say(output)
		case output := <-stdout:
			ui.Say(output)
		case exitStatus := <-exit:
			log.Printf("shell provisioner exited with status %d", exitStatus)
			break
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel first.
	for output := range stdout {
		ui.Say(output)
	}

	for output := range stderr {
		ui.Say(output)
	}
}
