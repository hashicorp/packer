// This package implements a provisioner for Packer that executes
// Puppet on the remote machine connecting to a Puppet master.
package puppetserver

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	
	// The command used to execute Puppet.
	ExecuteCommand string `mapstructure:"execute_command"`

	// Additional facts to set when executing Puppet
	Facter map[string]string

	// A path to the client certificate
	ClientCertPath string `mapstructure:"client_cert_path"`

	// A path to a directory containing the client private keys
	ClientPrivateKeyPath string `mapstructure:"client_private_key_path"`

	// The hostname of the Puppet node.
	PuppetNode string `mapstructure:"puppet_node"`

	// The hostname of the Puppet server.
	PuppetServer string `mapstructure:"puppet_server"`

	// Additional options to be passed to `puppet agent`.
	Options string `mapstructure:"options"`

	// If true, `sudo` will NOT be used to execute Puppet.
	PreventSudo bool `mapstructure:"prevent_sudo"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_dir"`

	// If true, packer will ignore all exit-codes from a puppet run
	IgnoreExitCodes bool `mapstructure:"ignore_exit_codes"`
}

type Provisioner struct {
	config Config
}

type ExecuteTemplate struct {
	FacterVars           string
	ClientCertPath       string
	ClientPrivateKeyPath string
	PuppetNode           string
	PuppetServer         string
	Options              string
	Sudo                 bool
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}
	
	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = p.commandTemplate()
	}
	
	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-puppet-server"
	}

	var errs *packer.MultiError
	if p.config.ClientCertPath != "" {
		info, err := os.Stat(p.config.ClientCertPath)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("client_cert_dir is invalid: %s", err))
		} else if !info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("client_cert_dir must point to a directory"))
		}
	}

	if p.config.ClientPrivateKeyPath != "" {
		info, err := os.Stat(p.config.ClientPrivateKeyPath)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("client_private_key_dir is invalid: %s", err))
		} else if !info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("client_private_key_dir must point to a directory"))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Puppet...")
	ui.Message("Creating Puppet staging directory...")
	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	// Upload client cert dir if set
	remoteClientCertPath := ""
	if p.config.ClientCertPath != "" {
		ui.Message(fmt.Sprintf(
			"Uploading client cert from: %s", p.config.ClientCertPath))
		remoteClientCertPath = fmt.Sprintf("%s/certs", p.config.StagingDir)
		err := p.uploadDirectory(ui, comm, remoteClientCertPath, p.config.ClientCertPath)
		if err != nil {
			return fmt.Errorf("Error uploading client cert: %s", err)
		}
	}

	// Upload client cert dir if set
	remoteClientPrivateKeyPath := ""
	if p.config.ClientPrivateKeyPath != "" {
		ui.Message(fmt.Sprintf(
			"Uploading client private keys from: %s", p.config.ClientPrivateKeyPath))
		remoteClientPrivateKeyPath = fmt.Sprintf("%s/private_keys", p.config.StagingDir)
		err := p.uploadDirectory(ui, comm, remoteClientPrivateKeyPath, p.config.ClientPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("Error uploading client private keys: %s", err)
		}
	}

	// Compile the facter variables
	facterVars := make([]string, 0, len(p.config.Facter))
	for k, v := range p.config.Facter {
		facterVars = append(facterVars, fmt.Sprintf("FACTER_%s='%s'", k, v))
	}

	// Execute Puppet
	p.config.ctx.Data = &ExecuteTemplate{
		FacterVars:           strings.Join(facterVars, " "),
		ClientCertPath:       remoteClientCertPath,
		ClientPrivateKeyPath: remoteClientPrivateKeyPath,
		PuppetNode:           p.config.PuppetNode,
		PuppetServer:         p.config.PuppetServer,
		Options:              p.config.Options,
		Sudo:                 !p.config.PreventSudo,
	}
	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	ui.Message(fmt.Sprintf("Running Puppet: %s", command))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 && cmd.ExitStatus != 2 && !p.config.IgnoreExitCodes {
		return fmt.Errorf("Puppet exited with a non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
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

func (p *Provisioner) uploadDirectory(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	// Make sure there is a trailing "/" so that the directory isn't
	// created on the other side.
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	return comm.UploadDir(dst, src, nil)
}

func (p *Provisioner) commandTemplate() string {
	return "{{.FacterVars}} {{if .Sudo}} sudo -E {{end}}" +
		"puppet agent --onetime --no-daemonize " +
		"{{if ne .PuppetServer \"\"}}--server='{{.PuppetServer}}' {{end}}" +
		"{{if ne .Options \"\"}}{{.Options}} {{end}}" +
		"{{if ne .PuppetNode \"\"}}--certname={{.PuppetNode}} {{end}}" +
		"{{if ne .ClientCertPath \"\"}}--certdir='{{.ClientCertPath}}' {{end}}" +
		"{{if ne .ClientPrivateKeyPath \"\"}}--privatekeydir='{{.ClientPrivateKeyPath}}' {{end}}" +
		"--detailed-exitcodes"
}
