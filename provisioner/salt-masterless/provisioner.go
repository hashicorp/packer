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
	common.PackerConfig `mapstructure:",squash"`

	// If true, run the salt-bootstrap script
	SkipBootstrap bool   `mapstructure:"skip_bootstrap"`
	BootstrapArgs string `mapstructure:"bootstrap_args"`

	// Local path to the minion config
	MinionConfig string `mapstructure:"minion_config"`

	// Local path to the salt state tree
	LocalStateTree string `mapstructure:"local_state_tree"`

	// Where files will be copied before moving to the /srv/salt directory
	TempConfigDir string `mapstructure:"temp_config_dir"`

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	if p.config.TempConfigDir == "" {
		p.config.TempConfigDir = DefaultTempConfigDir
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"bootstrap_args":   &p.config.BootstrapArgs,
		"minion_config":    &p.config.MinionConfig,
		"local_state_tree": &p.config.LocalStateTree,
		"temp_config_dir":  &p.config.TempConfigDir,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if p.config.LocalStateTree != "" {
		if _, err := os.Stat(p.config.LocalStateTree); err != nil {
			errs = packer.MultiErrorAppend(errs,
				errors.New("local_state_tree must exist and be accessible"))
		}
	}

	if p.config.MinionConfig != "" {
		if _, err := os.Stat(p.config.MinionConfig); err != nil {
			errs = packer.MultiErrorAppend(errs,
				errors.New("minion_config must exist and be accessible"))
		}
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

	if p.config.MinionConfig != "" {
		ui.Message(fmt.Sprintf("Uploading minion config: %s", p.config.MinionConfig))
		err := uploadMinionConfig(comm, "/etc/salt/minion", p.config.MinionConfig)
		if err != nil {
			return err
		}
	}

	if err = UploadLocalDirectory(p.config.LocalStateTree, p.config.TempConfigDir, comm, ui); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	ui.Message(fmt.Sprintf("Creating remote directory: %s", p.config.TempConfigDir))
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Error creating remote salt state directory: %s", err)
	}

	ui.Message(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	if err = UploadLocalDirectory(p.config.LocalStateTree, p.config.TempConfigDir, comm, ui); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	ui.Message(fmt.Sprintf("Moving %s to /srv/salt", p.config.TempConfigDir))
	cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s /srv/salt", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Unable to move %s to /srv/salt: %d", p.config.TempConfigDir, err)
	}

	ui.Message("Running highstate")
	cmd = &packer.RemoteCmd{Command: "sudo salt-call --local state.highstate -l info"}
	if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Error executing highstate: %s", err)
	}

	return nil
}

func UploadLocalDirectory(localDir string, remoteDir string, comm packer.Communicator, ui packer.Ui) (err error) {
	visitPath := func(localPath string, f os.FileInfo, err error) (err2 error) {
		localRelPath := strings.Replace(localPath, localDir, "", 1)
		localRelPath = strings.Replace(localRelPath, "\\", "/", -1)
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

func uploadMinionConfig(comm packer.Communicator, dst string, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening minion config: %s", err)
	}
	defer f.Close()

	if err = comm.Upload(dst, f); err != nil {
		return fmt.Errorf("Error uploading minion config: %s", err)
	}

	return nil
}
