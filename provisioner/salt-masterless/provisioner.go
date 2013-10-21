// This package implements a provisioner for Packer that executes a
// saltstack highstate within the remote machine
package saltmasterless

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
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

	// Local path to the salt pillar roots
	LocalPillarRoots string `mapstructure:"local_pillar_roots"`

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
		"bootstrap_args":     &p.config.BootstrapArgs,
		"minion_config":      &p.config.MinionConfig,
		"local_state_tree":   &p.config.LocalStateTree,
		"local_pillar_roots": &p.config.LocalPillarRoots,
		"temp_config_dir":    &p.config.TempConfigDir,
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

	if p.config.LocalPillarRoots != "" {
		if _, err := os.Stat(p.config.LocalPillarRoots); err != nil {
			errs = packer.MultiErrorAppend(errs,
				errors.New("local_pillar_roots must exist and be accessible"))
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

	ui.Message(fmt.Sprintf("Creating remote directory: %s", p.config.TempConfigDir))
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p %s", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Error creating remote salt state directory: %s", err)
	}

	if p.config.MinionConfig != "" {
		ui.Message(fmt.Sprintf("Uploading minion config: %s", p.config.MinionConfig))
		if err = uploadMinionConfig(comm, fmt.Sprintf("%s/minion", p.config.TempConfigDir), p.config.MinionConfig); err != nil {
			return fmt.Errorf("Error uploading local minion config file to remote: %s", err)
		}

		ui.Message(fmt.Sprintf("Moving %s/minion to /etc/salt/minion", p.config.TempConfigDir))
		cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s/minion /etc/salt/minion", p.config.TempConfigDir)}
		if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
			if err == nil {
				err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
			}

			return fmt.Errorf("Unable to move %s/minion to /etc/salt/minion: %d", p.config.TempConfigDir, err)
		}
	}

	ui.Message(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	if err = comm.UploadDir(fmt.Sprintf("%s/states", p.config.TempConfigDir),
		p.config.LocalStateTree, []string{".git"}); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	ui.Message(fmt.Sprintf("Moving %s/states to /srv/salt", p.config.TempConfigDir))
	cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s/states /srv/salt", p.config.TempConfigDir)}
	if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Unable to move %s/states to /srv/salt: %d", p.config.TempConfigDir, err)
	}

	if p.config.LocalPillarRoots != "" {
		ui.Message(fmt.Sprintf("Uploading local pillar roots: %s", p.config.LocalPillarRoots))
		if err = comm.UploadDir(fmt.Sprintf("%s/pillar", p.config.TempConfigDir),
			p.config.LocalPillarRoots, []string{".git"}); err != nil {
			return fmt.Errorf("Error uploading local pillar roots to remote: %s", err)
		}

		ui.Message(fmt.Sprintf("Moving %s/pillar to /srv/pillar", p.config.TempConfigDir))
		cmd = &packer.RemoteCmd{Command: fmt.Sprintf("sudo mv %s/pillar /srv/pillar", p.config.TempConfigDir)}
		if err = cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
			if err == nil {
				err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
			}

			return fmt.Errorf("Unable to move %s/pillar to /srv/pillar: %d", p.config.TempConfigDir, err)
		}
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

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
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
