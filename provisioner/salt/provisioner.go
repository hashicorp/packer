// This package implements a provisioner for Packer that executes a
// saltstack highstate within the remote machine
package salt

import (
	"errors"
	"fmt"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var Ui packer.Ui

const DefaultTempConfigDir = "/tmp/salt"

type config struct {
	// If true, run the salt-bootstrap script
	SkipBootstrap bool   `mapstructure:"skip_bootstrap"`
	BootstrapArgs string `mapstructure:"bootstrap_args"`

	// Local path to the salt state tree
	LocalStateTree string `mapstructure:"local_state_tree"`

	// Where files will be copied before moving to the /srv/salt directory
	TempConfigDir string `mapstructure:"temp_config_dir"`
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	var md mapstructure.Metadata
	decoderConfig := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &p.config,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	for _, raw := range raws {
		err := decoder.Decode(raw)
		if err != nil {
			return err
		}
	}

	// Accumulate any errors
	errs := make([]error, 0)

	// Unused keys are errors
	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			if unused != "type" && !strings.HasPrefix(unused, "packer_") {
				errs = append(
					errs, fmt.Errorf("Unknown configuration key: %s", unused))
			}
		}
	}

	if p.config.LocalStateTree == "" {
		errs = append(errs, errors.New("Please specify a local_state_tree"))
	}

	if p.config.TempConfigDir == "" {
		p.config.TempConfigDir = DefaultTempConfigDir
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	var err error
	Ui = ui

	if !p.config.SkipBootstrap {
		cmd := fmt.Sprintf("wget -O - http://bootstrap.saltstack.org | sudo sh -s %s", p.config.BootstrapArgs)
		Ui.Say(fmt.Sprintf("Installing Salt with command %s", cmd))
		if err = ExecuteCommand(cmd, comm); err != nil {
			return fmt.Errorf("Unable to install Salt: %d", err)
		}
	}

	Ui.Say(fmt.Sprintf("Creating remote directory: %s", p.config.TempConfigDir))
	if err = ExecuteCommand(fmt.Sprintf("mkdir -p %s", p.config.TempConfigDir), comm); err != nil {
		return fmt.Errorf("Error creating remote salt state directory: %s", err)
	}

	Ui.Say(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	if err = UploadLocalDirectory(p.config.LocalStateTree, p.config.TempConfigDir, comm); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	Ui.Say(fmt.Sprintf("Moving %s to /srv/salt", p.config.TempConfigDir))
	if err = ExecuteCommand(fmt.Sprintf("sudo mv %s /srv/salt", p.config.TempConfigDir), comm); err != nil {
		return fmt.Errorf("Unable to move %s to /srv/salt: %d", p.config.TempConfigDir, err)
	}

	Ui.Say("Running highstate")
	if err = ExecuteCommand("sudo salt-call --local state.highstate -l info", comm); err != nil {
		return fmt.Errorf("Error executing highstate: %s", err)
	}

	Ui.Say("Removing /srv/salt")
	if err = ExecuteCommand("sudo rm -r /srv/salt", comm); err != nil {
		return fmt.Errorf("Unable to remove /srv/salt: %d", err)
	}

	return nil
}

func UploadLocalDirectory(localDir string, remoteDir string, comm packer.Communicator) (err error) {
	visitPath := func(localPath string, f os.FileInfo, err error) (err2 error) {
		localRelPath := strings.Replace(localPath, localDir, "", 1)
		remotePath := fmt.Sprintf("%s%s", remoteDir, localRelPath)
		if f.IsDir() && f.Name() == ".git" {
			return filepath.SkipDir
		}
		if f.IsDir() {
			// Make remote directory
			err = ExecuteCommand(fmt.Sprintf("mkdir -p %s", remotePath), comm)
			if err != nil {
				return err
			}
		} else {
			// Upload file to existing directory
			file, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("Error opening file: %s", err)
			}
			defer file.Close()

			Ui.Say(fmt.Sprintf("Uploading file %s: %s", localPath, remotePath))
			err = comm.Upload(remotePath, file)
			if err != nil {
				return fmt.Errorf("Error uploading file: %s", err)
			}
		}
		return
	}

	err = filepath.Walk(localDir, visitPath)
	if err != nil {
		return fmt.Errorf("Error uploading local directory %s: %s", localDir, err)
	}

	return nil
}

func ExecuteCommand(command string, comm packer.Communicator) (err error) {
	// Setup the remote command
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = command
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
			Ui.Message(strings.TrimSpace(output))
		case output := <-stdoutChan:
			Ui.Message(strings.TrimSpace(output))
		case exitStatus := <-exitChan:
			log.Printf("Salt provisioner exited with status %d", exitStatus)

			if exitStatus != 0 {
				return fmt.Errorf("Command exited with non-zero exit status: %d", exitStatus)
			}

			break OutputLoop
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel first.
	for output := range stdoutChan {
		Ui.Message(output)
	}

	for output := range stderrChan {
		Ui.Message(output)
	}

	return nil
}
