//go:generate mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that executes a
// saltstack state within the remote machine
package saltmasterless

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// If true, run the salt-bootstrap script
	SkipBootstrap bool   `mapstructure:"skip_bootstrap"`
	BootstrapArgs string `mapstructure:"bootstrap_args"`

	DisableSudo bool `mapstructure:"disable_sudo"`

	// Custom state to run instead of highstate
	CustomState string `mapstructure:"custom_state"`

	// Local path to the minion config
	MinionConfig string `mapstructure:"minion_config"`

	// Local path to the minion grains
	GrainsFile string `mapstructure:"grains_file"`

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

	// Don't exit packer if salt-call returns an error code
	NoExitOnFailure bool `mapstructure:"no_exit_on_failure"`

	// Set the logging level for the salt-call run
	LogLevel string `mapstructure:"log_level"`

	// Arguments to pass to salt-call
	SaltCallArgs string `mapstructure:"salt_call_args"`

	// Directory containing salt-call
	SaltBinDir string `mapstructure:"salt_bin_dir"`

	// Command line args passed onto salt-call
	CmdArgs string ""

	// The Guest OS Type (unix or windows)
	GuestOSType string `mapstructure:"guest_os_type"`

	ctx interpolate.Context
}

type Provisioner struct {
	config            Config
	guestOSTypeConfig guestOSTypeConfig
	guestCommands     *provisioner.GuestCommands
}

type guestOSTypeConfig struct {
	tempDir           string
	stateRoot         string
	pillarRoot        string
	configDir         string
	bootstrapFetchCmd string
	bootstrapRunCmd   string
}

var guestOSTypeConfigs = map[string]guestOSTypeConfig{
	provisioner.UnixOSType: {
		configDir:         "/etc/salt",
		tempDir:           "/tmp/salt",
		stateRoot:         "/srv/salt",
		pillarRoot:        "/srv/pillar",
		bootstrapFetchCmd: "curl -L https://bootstrap.saltstack.com -o /tmp/install_salt.sh || wget -O /tmp/install_salt.sh https://bootstrap.saltstack.com",
		bootstrapRunCmd:   "sh /tmp/install_salt.sh",
	},
	provisioner.WindowsOSType: {
		configDir:         "C:/salt/conf",
		tempDir:           "C:/Windows/Temp/salt/",
		stateRoot:         "C:/salt/state",
		pillarRoot:        "C:/salt/pillar/",
		bootstrapFetchCmd: "powershell Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/saltstack/salt-bootstrap/stable/bootstrap-salt.ps1' -OutFile 'C:/Windows/Temp/bootstrap-salt.ps1'",
		bootstrapRunCmd:   "Powershell C:/Windows/Temp/bootstrap-salt.ps1",
	},
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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

	if p.config.GuestOSType == "" {
		p.config.GuestOSType = provisioner.DefaultOSType
	} else {
		p.config.GuestOSType = strings.ToLower(p.config.GuestOSType)
	}

	var ok bool
	p.guestOSTypeConfig, ok = guestOSTypeConfigs[p.config.GuestOSType]
	if !ok {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	p.guestCommands, err = provisioner.NewGuestCommands(p.config.GuestOSType, !p.config.DisableSudo)
	if err != nil {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	if p.config.TempConfigDir == "" {
		p.config.TempConfigDir = p.guestOSTypeConfig.tempDir
	}

	var errs *packer.MultiError

	// require a salt state tree
	err = validateDirConfig(p.config.LocalStateTree, "local_state_tree", true)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	err = validateDirConfig(p.config.LocalPillarRoots, "local_pillar_roots", false)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	err = validateFileConfig(p.config.MinionConfig, "minion_config", false)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if p.config.MinionConfig != "" && (p.config.RemoteStateTree != "" || p.config.RemotePillarRoots != "") {
		errs = packer.MultiErrorAppend(errs,
			errors.New("remote_state_tree and remote_pillar_roots only apply when minion_config is not used"))
	}

	err = validateFileConfig(p.config.GrainsFile, "grains_file", false)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	// build the command line args to pass onto salt
	var cmd_args bytes.Buffer

	if p.config.CustomState == "" {
		cmd_args.WriteString(" state.highstate")
	} else {
		cmd_args.WriteString(" state.sls ")
		cmd_args.WriteString(p.config.CustomState)
	}

	if p.config.MinionConfig == "" {
		// pass --file-root and --pillar-root if no minion_config is supplied
		if p.config.RemoteStateTree != "" {
			cmd_args.WriteString(" --file-root=")
			cmd_args.WriteString(p.config.RemoteStateTree)
		} else {
			cmd_args.WriteString(" --file-root=")
			cmd_args.WriteString(p.guestOSTypeConfig.stateRoot)
		}
		if p.config.RemotePillarRoots != "" {
			cmd_args.WriteString(" --pillar-root=")
			cmd_args.WriteString(p.config.RemotePillarRoots)
		} else {
			cmd_args.WriteString(" --pillar-root=")
			cmd_args.WriteString(p.guestOSTypeConfig.pillarRoot)
		}
	}

	if !p.config.NoExitOnFailure {
		cmd_args.WriteString(" --retcode-passthrough")
	}

	if p.config.LogLevel == "" {
		cmd_args.WriteString(" -l info")
	} else {
		cmd_args.WriteString(" -l ")
		cmd_args.WriteString(p.config.LogLevel)
	}

	if p.config.SaltCallArgs != "" {
		cmd_args.WriteString(" ")
		cmd_args.WriteString(p.config.SaltCallArgs)
	}

	p.config.CmdArgs = cmd_args.String()

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	var err error
	var src, dst string

	ui.Say("Provisioning with Salt...")
	if !p.config.SkipBootstrap {
		cmd := &packer.RemoteCmd{
			// Fallback on wget if curl failed for any reason (such as not being installed)
			Command: fmt.Sprintf(p.guestOSTypeConfig.bootstrapFetchCmd),
		}
		ui.Message(fmt.Sprintf("Downloading saltstack bootstrap to /tmp/install_salt.sh"))
		if err = cmd.RunWithUi(ctx, comm, ui); err != nil {
			return fmt.Errorf("Unable to download Salt: %s", err)
		}
		cmd = &packer.RemoteCmd{
			Command: fmt.Sprintf("%s %s", p.sudo(p.guestOSTypeConfig.bootstrapRunCmd), p.config.BootstrapArgs),
		}
		ui.Message(fmt.Sprintf("Installing Salt with command %s", cmd.Command))
		if err = cmd.RunWithUi(ctx, comm, ui); err != nil {
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
		ui.Message(fmt.Sprintf("Make sure directory %s exists", p.guestOSTypeConfig.configDir))
		if err := p.createDir(ui, comm, p.guestOSTypeConfig.configDir); err != nil {
			return fmt.Errorf("Error creating remote salt configuration directory: %s", err)
		}
		src = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "minion"))
		dst = filepath.ToSlash(filepath.Join(p.guestOSTypeConfig.configDir, "minion"))
		if err = p.moveFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Unable to move %s/minion to %s/minion: %s", p.config.TempConfigDir, p.guestOSTypeConfig.configDir, err)
		}
	}

	if p.config.GrainsFile != "" {
		ui.Message(fmt.Sprintf("Uploading grains file: %s", p.config.GrainsFile))
		src = p.config.GrainsFile
		dst = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "grains"))
		if err = p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading local grains file to remote: %s", err)
		}

		// move grains file into /etc/salt
		ui.Message(fmt.Sprintf("Make sure directory %s exists", p.guestOSTypeConfig.configDir))
		if err := p.createDir(ui, comm, p.guestOSTypeConfig.configDir); err != nil {
			return fmt.Errorf("Error creating remote salt configuration directory: %s", err)
		}
		src = filepath.ToSlash(filepath.Join(p.config.TempConfigDir, "grains"))
		dst = filepath.ToSlash(filepath.Join(p.guestOSTypeConfig.configDir, "grains"))
		if err = p.moveFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Unable to move %s/grains to %s/grains: %s", p.config.TempConfigDir, p.guestOSTypeConfig.configDir, err)
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
	if p.config.RemoteStateTree != "" {
		dst = p.config.RemoteStateTree
	} else {
		dst = p.guestOSTypeConfig.stateRoot
	}

	if err = p.statPath(ui, comm, dst); err != nil {
		if err = p.removeDir(ui, comm, dst); err != nil {
			return fmt.Errorf("Unable to clear salt tree: %s", err)
		}
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
		if p.config.RemotePillarRoots != "" {
			dst = p.config.RemotePillarRoots
		} else {
			dst = p.guestOSTypeConfig.pillarRoot
		}

		if err = p.statPath(ui, comm, dst); err != nil {
			if err = p.removeDir(ui, comm, dst); err != nil {
				return fmt.Errorf("Unable to clear pillar root: %s", err)
			}
		}

		if err = p.moveFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Unable to move %s/pillar to %s: %s", p.config.TempConfigDir, dst, err)
		}
	}

	ui.Message(fmt.Sprintf("Running: salt-call --local %s", p.config.CmdArgs))
	cmd := &packer.RemoteCmd{Command: p.sudo(fmt.Sprintf("%s --local %s", filepath.Join(p.config.SaltBinDir, "salt-call"), p.config.CmdArgs))}
	if err = cmd.RunWithUi(ctx, comm, ui); err != nil || cmd.ExitStatus() != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus())
		}

		return fmt.Errorf("Error executing salt-call: %s", err)
	}

	return nil
}

// Prepends sudo to supplied command if config says to
func (p *Provisioner) sudo(cmd string) string {
	if p.config.DisableSudo || (p.config.GuestOSType == provisioner.WindowsOSType) {
		return cmd
	}

	return "sudo " + cmd
}

func validateDirConfig(path string, name string, required bool) error {
	if required && path == "" {
		return fmt.Errorf("%s cannot be empty", name)
	} else if required == false && path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: path '%s' is invalid: %s", name, path, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s: path '%s' must point to a directory", name, path)
	}
	return nil
}

func validateFileConfig(path string, name string, required bool) error {
	if required == true && path == "" {
		return fmt.Errorf("%s cannot be empty", name)
	} else if required == false && path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: path '%s' is invalid: %s", name, path, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: path '%s' must point to a file", name, path)
	}
	return nil
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening: %s", err)
	}
	defer f.Close()

	_, temp_dst := filepath.Split(dst)

	if err = comm.Upload(temp_dst, f, nil); err != nil {
		return fmt.Errorf("Error uploading %s: %s", src, err)
	}

	p.moveFile(ui, comm, dst, temp_dst)

	return nil
}

func (p *Provisioner) moveFile(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	ctx := context.TODO()

	ui.Message(fmt.Sprintf("Moving %s to %s", src, dst))
	cmd := &packer.RemoteCmd{
		Command: p.sudo(p.guestCommands.MovePath(src, dst)),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil || cmd.ExitStatus() != 0 {
		if err == nil {
			err = fmt.Errorf("Bad exit status: %d", cmd.ExitStatus())
		}

		return fmt.Errorf("Unable to move %s to %s: %s", src, dst, err)
	}
	return nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: p.guestCommands.CreateDir(dir),
	}
	ctx := context.TODO()
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}
	return nil
}

func (p *Provisioner) statPath(ui packer.Ui, comm packer.Communicator, path string) error {
	ctx := context.TODO()
	ui.Message(fmt.Sprintf("Verifying Path: %s", path))
	cmd := &packer.RemoteCmd{
		Command: p.guestCommands.StatPath(path),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ctx := context.TODO()
	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: p.guestCommands.RemoveDir(dir),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}
	return nil
}

func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string, ignore []string) error {
	_, temp_dst := filepath.Split(dst)
	if err := comm.UploadDir(temp_dst, src, ignore); err != nil {
		return err
	}
	return p.moveFile(ui, comm, dst, temp_dst)
}
