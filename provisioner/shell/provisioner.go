// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"fmt"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"strings"
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

func (p *Provisioner) Prepare(raws ...interface{}) error {
	// TODO: errors
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}

	return nil
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

	// Setup the remote command
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = fmt.Sprintf("chmod +x %s && %s", p.config.RemotePath, p.config.RemotePath)
	cmd.Stdout = stdout_w
	cmd.Stderr = stderr_w

	log.Printf("Executing command: %s", cmd.Command)
	err = comm.Start(&cmd)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed executing command: %s", err))
		return
	}

	exitChan := make(chan int, 1)
	stdoutChan := iochan.DelimReader(stdout_r, '\n')
	stderrChan := iochan.DelimReader(stderr_r, '\n')

	go func() {
		defer stdout_w.Close()
		defer stderr_w.Close()

		cmd.Wait()
		exitChan <- cmd.ExitStatus
	}()

OutputLoop:
	for {
		select {
		case output := <-stderrChan:
			ui.Message(strings.TrimSpace(output))
		case output := <-stdoutChan:
			ui.Message(strings.TrimSpace(output))
		case exitStatus := <-exitChan:
			log.Printf("shell provisioner exited with status %d", exitStatus)
			break OutputLoop
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel first.
	for output := range stdoutChan {
		ui.Message(output)
	}

	for output := range stderrChan {
		ui.Message(output)
	}
}
