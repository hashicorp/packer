// This package implements a provisioner for Packer that executes a
// saltstack highstate within the remote machine
package salt

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
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
		cmd := &packer.RemoteCmd{
			Command: fmt.Sprintf("wget -O - http://bootstrap.saltstack.org | sudo sh -s %s", p.config.BootstrapArgs),
		}
		Ui.Say(fmt.Sprintf("Installing Salt with command %s", cmd))
		if err = cmd.StartWithUi(comm, ui); err != nil {
			return fmt.Errorf("Unable to install Salt: %d", err)
		}
	}

	Ui.Say(fmt.Sprintf("Creating remote directory: %s", p.config.TempConfigDir))
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Error creating remote salt state directory: %s", err)
	}

	Ui.Say(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	if err = UploadLocalDirectory(p.config.LocalStateTree, p.config.TempConfigDir, comm); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	Ui.Say(fmt.Sprintf("Moving %s to /srv/salt", p.config.TempConfigDir))
	cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s /srv/salt", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Unable to move %s to /srv/salt: %d", p.config.TempConfigDir, err)
	}

	Ui.Say("Running highstate")
	cmd = &packer.RemoteCmd{Command: "sudo salt-call --local state.highstate -l info"}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Error executing highstate: %s", err)
	}

	Ui.Say("Removing /srv/salt")
	cmd = &packer.RemoteCmd{Command: "sudo rm -r /srv/salt"}
	if err = cmd.StartWithUi(comm, ui); err != nil {
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
			cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", remotePath)}
			if err = cmd.StartWithUi(comm, Ui); err != nil {
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
			if err = comm.Upload(remotePath, file); err != nil {
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
