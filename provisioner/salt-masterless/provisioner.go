// This package implements a provisioner for Packer that executes a
// saltstack highstate within the remote machine
package saltmasterless

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
	"path/filepath"
	"strings"
)

const DefaultTempConfigDir = "/tmp/salt"

type Config struct {
	// If true, run the salt-bootstrap script
	SkipBootstrap bool   `mapstructure:"skip_bootstrap"`
	BootstrapArgs string `mapstructure:"bootstrap_args"`

	// Local path to the salt state tree
	LocalStateTree string `mapstructure:"local_state_tree"`

	// Where files will be copied before moving to the /srv/salt directory
	TempConfigDir string `mapstructure:"temp_config_dir"`
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	if p.config.TempConfigDir == "" {
		p.config.TempConfigDir = DefaultTempConfigDir
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if p.config.LocalStateTree == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Please specify a local_state_tree"))
	} else if _, err := os.Stat(p.config.LocalStateTree); err != nil {
		errs = packer.MultiErrorAppend(errs,
			errors.New("local_state_tree must exist and be accessible"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	var err error

	ui.Say("Provisioning with Salt...")
	if !p.config.SkipBootstrap {
		cmd := &packer.RemoteCmd{
			Command: fmt.Sprintf("wget -O - http://bootstrap.saltstack.org | sudo sh -s %s", p.config.BootstrapArgs),
		}
		ui.Message(fmt.Sprintf("Installing Salt with command %s", cmd))
		if err = cmd.StartWithUi(comm, ui); err != nil {
			return fmt.Errorf("Unable to install Salt: %d", err)
		}
	}

	ui.Message(fmt.Sprintf("Creating remote directory: %s", p.config.TempConfigDir))
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Error creating remote salt state directory: %s", err)
	}

	ui.Message(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	if err = UploadLocalDirectory(p.config.LocalStateTree, p.config.TempConfigDir, comm, ui); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	ui.Message(fmt.Sprintf("Moving %s to /srv/salt", p.config.TempConfigDir))
	cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s /srv/salt", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Unable to move %s to /srv/salt: %d", p.config.TempConfigDir, err)
	}

	ui.Message("Running highstate")
	cmd = &packer.RemoteCmd{Command: "sudo salt-call --local state.highstate -l info"}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Error executing highstate: %s", err)
	}

	ui.Message("Removing /srv/salt")
	cmd = &packer.RemoteCmd{Command: "sudo rm -r /srv/salt"}
	if err = cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf("Unable to remove /srv/salt: %d", err)
	}

	return nil
}

func UploadLocalDirectory(localDir string, remoteDir string, comm packer.Communicator, ui packer.Ui) (err error) {
	visitPath := func(localPath string, f os.FileInfo, err error) (err2 error) {
		localRelPath := strings.Replace(localPath, localDir, "", 1)
		remotePath := fmt.Sprintf("%s%s", remoteDir, localRelPath)
		if f.IsDir() && f.Name() == ".git" {
			return filepath.SkipDir
		}
		if f.IsDir() {
			// Make remote directory
			cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", remotePath)}
			if err = cmd.StartWithUi(comm, ui); err != nil {
				return err
			}
		} else {
			// Upload file to existing directory
			file, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("Error opening file: %s", err)
			}
			defer file.Close()

			ui.Message(fmt.Sprintf("Uploading file %s: %s", localPath, remotePath))
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
