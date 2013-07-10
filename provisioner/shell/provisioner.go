// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

	// The shebang value used when running inline scripts.
	InlineShebang string `mapstructure:"inline_shebang"`

	// The local path of the shell script to upload and execute.
	Script string

	// An array of multiple scripts to run.
	Scripts []string

	// An array of environment variables that will be injected before
	// your command(s) are executed.
	Vars []string `mapstructure:"environment_vars"`

	// The remote path where the local shell script will be uploaded to.
	// This should be set to a writable file that is in a pre-existing directory.
	RemotePath string `mapstructure:"remote_path"`

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand string `mapstructure:"execute_command"`
}

type Provisioner struct {
	config config
}

type ExecuteCommandTemplate struct {
	Vars string
	Path string
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "chmod +x {{.Path}}; {{.Vars}} {{.Path}}"
	}

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.InlineShebang == "" {
		p.config.InlineShebang = "/bin/sh"
	}

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}

	if p.config.Scripts == nil {
		p.config.Scripts = make([]string, 0)
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	errs := make([]error, 0)

	if p.config.Script != "" && len(p.config.Scripts) > 0 {
		errs = append(errs, errors.New("Only one of script or scripts can be specified."))
	}

	if p.config.Script != "" {
		p.config.Scripts = []string{p.config.Script}
	}

	if len(p.config.Scripts) == 0 && p.config.Inline == nil {
		errs = append(errs, errors.New("Either a script file or inline script must be specified."))
	} else if len(p.config.Scripts) > 0 && p.config.Inline != nil {
		errs = append(errs, errors.New("Only a script file or an inline script can be specified, not both."))
	}

	for _, path := range p.config.Scripts {
		if _, err := os.Stat(path); err != nil {
			errs = append(errs, fmt.Errorf("Bad script '%s': %s", path, err))
		}
	}

	// Do a check for bad environment variables, such as '=foo', 'foobar'
	for _, kv := range p.config.Vars {
		vs := strings.Split(kv, "=")
		if len(vs) != 2 || vs[0] == "" {
			errs = append(errs, fmt.Errorf("Environment variable not in format 'key=value': %s", kv))
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if p.config.Inline != nil {
		tf, err := ioutil.TempFile("", "packer-shell")
		if err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		scripts = append(scripts, tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		writer.WriteString(fmt.Sprintf("#!%s\n", p.config.InlineShebang))
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return fmt.Errorf("Error preparing shell script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}

		tf.Close()
	}

	for _, path := range scripts {
		ui.Say(fmt.Sprintf("Provisioning with shell script: %s", path))

		log.Printf("Opening %s for reading", path)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening shell script: %s", err)
		}
		defer f.Close()

		log.Printf("Uploading %s => %s", path, p.config.RemotePath)
		err = comm.Upload(p.config.RemotePath, f)
		if err != nil {
			return fmt.Errorf("Error uploading shell script: %s", err)
		}

		// Close the original file since we copied it
		f.Close()

		// Flatten the environment variables
		flattendVars := strings.Join(p.config.Vars, " ")

		// Compile the command
		var command bytes.Buffer
		t := template.Must(template.New("command").Parse(p.config.ExecuteCommand))
		t.Execute(&command, &ExecuteCommandTemplate{flattendVars, p.config.RemotePath})

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
			return fmt.Errorf("Failed executing command: %s", err)
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

				if exitStatus != 0 {
					return fmt.Errorf("Script exited with non-zero exit status: %d", exitStatus)
				}

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

	return nil
}
