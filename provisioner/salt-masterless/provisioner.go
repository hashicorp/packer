// This package implements a provisioner for Packer that executes a
// saltstack highstate within the remote machine
package saltmasterless

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const DefaultTempConfigDir = "/tmp/salt"
const DefaultStateTreeDir = "/srv/salt"
const DefaultPillarRootDir = "/srv/pillar"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// If true, run the salt-bootstrap script
	SkipBootstrap bool   `mapstructure:"skip_bootstrap"`
	BootstrapArgs string `mapstructure:"bootstrap_args"`

	DisableSudo bool `mapstructure:"disable_sudo"`

	// Local path to the minion config
	MinionConfig string `mapstructure:"minion_config"`

	// Local path to the salt state tree
	LocalStateTree string `mapstructure:"local_state_tree"`

	// Local path to the salt pillar roots
	LocalPillarRoots string `mapstructure:"local_pillar_roots"`

	// Remote path to the salt state tree
	RemoteStateTree string `mapstructure:"remote_state_tree"`

	// Remote path to the salt pillar roots
	RemotePillarRoots string `mapstructure:"remote_pillar_roots"`

	// Where files will be copied before moving to the /srv/salt directory
	TempConfigDir string `mapstructure:"temp_config_dir"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.TempConfigDir == "" {
		p.config.TempConfigDir = DefaultTempConfigDir
	}

	if p.config.RemoteStateTree == "" {
		p.config.RemoteStateTree = DefaultStateTreeDir
	}

	if p.config.RemotePillarRoots == "" {
		p.config.RemotePillarRoots = DefaultPillarRootDir
	}

	var errs *packer.MultiError

	// require a salt state tree
	if p.config.LocalStateTree == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("local_state_tree must be supplied"))
	} else {
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
	var src, dst string

	ui.Say("Provisioning with Salt...")
	if !p.config.SkipBootstrap {
		cmd := &packer.RemoteCmd{
			Command: fmt.Sprintf("curl -L https://bootstrap.saltstack.com -o /tmp/install_salt.sh"),
		}
		ui.Message(fmt.Sprintf("Downloading saltstack bootstrap to /tmp/install_salt.sh"))
		if err = cmd.StartWithUi(comm, ui); err != nil {
			return fmt.Errorf("Unable to download Salt: %s", err)
		}
		cmd = &packer.RemoteCmd{
			Command: fmt.Sprintf("%s /tmp/install_salt.sh %s", p.sudo("sh"), p.config.BootstrapArgs),
		}
		ui.Message(fmt.Sprintf("Installing Salt with command %s", cmd.Command))
		if err = cmd.StartWithUi(comm, ui); err != nil {
			return fmt.Errorf("Unable to install Salt: %s", err)
		}
	}

	ui.Message(fmt.Sprintf("Creating remote temporary directory: %s", p.config.TempConfigDir))
	if err := p.createDir(ui, comm, p.config.TempConfigDir); err != nil {
		return fmt.Errorf("Error creating remote temporary directory: %s", err)
	}

	if p.config.MinionConfig != "" {
		ui.Message(fmt.Sprintf("Uploading minion config: %s", p.config.MinionConfig))
		src = p.config.MinionConfig
		dst = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "minion"))
		if err = p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading local minion config file to remote: %s", err)
		}

		// move minion config into /etc/salt
		ui.Message(fmt.Sprintf("Make sure directory %s exists", "/etc/salt"))
		if err := p.createDir(ui, comm, "/etc/salt"); err != nil {
			return fmt.Errorf("Error creating remote salt configuration directory: %s", err)
		}
		src = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "minion"))
		dst = "/etc/salt/minion"
		if err = p.moveFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Unable to move %s/minion to /etc/salt/minion: %s", p.config.TempConfigDir, err)
		}
	}

	ui.Message(fmt.Sprintf("Uploading local state tree: %s", p.config.LocalStateTree))
	src = p.config.LocalStateTree
	dst = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "states"))
	if err = p.uploadDir(ui, comm, dst, src, []string{".git"}); err != nil {
		return fmt.Errorf("Error uploading local state tree to remote: %s", err)
	}

	// move state tree from temporary directory
	src = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "states"))
	dst = p.config.RemoteStateTree
	if err = p.removeDir(ui, comm, dst); err != nil {
		return fmt.Errorf("Unable to clear salt tree: %s", err)
	}
	if err = p.moveFile(ui, comm, dst, src); err != nil {
		return fmt.Errorf("Unable to move %s/states to %s: %s", p.config.TempConfigDir, dst, err)
	}

	if p.config.LocalPillarRoots != "" {
		ui.Message(fmt.Sprintf("Uploading local pillar roots: %s", p.config.LocalPillarRoots))
		src = p.config.LocalPillarRoots
		dst = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "pillar"))
		if err = p.uploadDir(ui, comm, dst, src, []string{".git"}); err != nil {
			return fmt.Errorf("Error uploading local pillar roots to remote: %s", err)
		}

		// move pillar root from temporary directory
		src = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "pillar"))
		dst = p.config.RemotePillarRoots
		if err = p.removeDir(ui, comm, dst); err != nil {
			return fmt.Errorf("Unable to clear pillat root: %s", err)
		}
		if err = p.moveFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Unable to move %s/pillar to %s: %s", p.config.TempConfigDir, dst, err)
		}
	}

	ui.Message("Running highstate")
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf(p.sudo("salt-call --local state.highstate --file-root=%s --pillar-root=%s -l info --retcode-passthrough"),p.config.RemoteStateTree, p.config.RemotePillarRoots)}
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

// Prepends sudo to supplied command if config says to
func (p *Provisioner) sudo(cmd string) string {
	if p.config.DisableSudo {
		return cmd
	}

	return "sudo " + cmd
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening: %s", err)
	}
	defer f.Close()

	if err = comm.Upload(dst, f, nil); err != nil {
		return fmt.Errorf("Error uploading %s: %s", src, err)
	}
	return nil
}

func (p *Provisioner) moveFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	ui.Message(fmt.Sprintf("Moving %s to %s", src, dst))
	cmd := &packer.RemoteCmd{Command: fmt.Sprintf(p.sudo("mv %s %s"), src, dst)}
	if err := cmd.StartWithUi(comm, ui); err != nil || cmd.ExitStatus != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus)
		}

		return fmt.Errorf("Unable to move %s to %s: %s", src, dst, err)
	}
	return nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("rm -rf '%s'", dir),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}
	return nil
}

func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string, ignore []string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	// Make sure there is a trailing "/" so that the directory isn't
	// created on the other side.
	if src[len(src)-1] != '/' {
		src = src + "/"
	}
	return comm.UploadDir(dst, src, ignore)
}
