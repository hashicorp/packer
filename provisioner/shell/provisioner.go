// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"errors"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

const DefaultRemotePath = "/tmp/script.sh"

type config struct {
	// An inline script to execute. Multiple strings are all executed
	// in the context of a single shell.
	Inline []string

	// The local path of the shell script to upload and execute.
	Path string

	// The remote path where the local shell script will be uploaded to.
	// This should be set to a writable file that is in a pre-existing directory.
	RemotePath string `mapstructure:"remote_path"`

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes.
	ExecuteCommand string `mapstructure:"execute_command"`
}

type Provisioner struct {
	config config
}

type ExecuteCommandTemplate struct {
	Path string
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "sh {{.Path}}"
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}

	errs := make([]error, 0)

	if p.config.Path == "" && p.config.Inline == nil {
		errs = append(errs, errors.New("Either a path or inline script must be specified."))
	} else if p.config.Path != "" && p.config.Inline != nil {
		errs = append(errs, errors.New("Only a path or an inline script can be specified, not both."))
	}

	if p.config.Path != "" {
		if _, err := os.Stat(p.config.Path); err != nil {
			errs = append(errs, fmt.Errorf("Bad script path: %s", err))
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	path := p.config.Path

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if p.config.Inline != nil {
		tf, err := ioutil.TempFile("", "packer-shell")
		if err != nil {
			ui.Error(fmt.Sprintf("Error preparing shell script: %s", err))
			return
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		path = tf.Name()

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command+"\n"); err != nil {
				ui.Error(fmt.Sprintf("Error preparing shell script: %s", err))
				return
			}
		}

		if err := writer.Flush(); err != nil {
			ui.Error(fmt.Sprintf("Error preparing shell script: %s", err))
			return
		}

		tf.Close()
	}

	ui.Say(fmt.Sprintf("Provisioning with shell script: %s", path))

	log.Printf("Opening %s for reading", path)
	f, err := os.Open(path)
	if err != nil {
		ui.Error(fmt.Sprintf("Error opening shell script: %s", err))
		return
	}

	log.Printf("Uploading %s => %s", path, p.config.RemotePath)
	err = comm.Upload(p.config.RemotePath, f)
	if err != nil {
		ui.Error(fmt.Sprintf("Error uploading shell script: %s", err))
		return
	}

	// Compile the command
	var command bytes.Buffer
	t := template.Must(template.New("command").Parse(p.config.ExecuteCommand))
	t.Execute(&command, &ExecuteCommandTemplate{p.config.RemotePath})

	// Setup the remote command
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = command.String()
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
