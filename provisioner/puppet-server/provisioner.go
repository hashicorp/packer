//go:generate mapstructure-to-hcl2 -type Config

// Package puppetserver implements a provisioner for Packer that executes
// Puppet on the remote machine connecting to a Puppet master.
package puppetserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// If true, staging directory is removed after executing puppet.
	CleanStagingDir bool `mapstructure:"clean_staging_directory"`

	// A path to the client certificate
	ClientCertPath string `mapstructure:"client_cert_path"`

	// A path to a directory containing the client private keys
	ClientPrivateKeyPath string `mapstructure:"client_private_key_path"`

	// The command used to execute Puppet.
	ExecuteCommand string `mapstructure:"execute_command"`

	// Additional argument to pass when executing Puppet.
	ExtraArguments []string `mapstructure:"extra_arguments"`

	// Additional facts to set when executing Puppet
	Facter map[string]string

	// The Guest OS Type (unix or windows)
	GuestOSType string `mapstructure:"guest_os_type"`

	// If true, packer will ignore all exit-codes from a puppet run
	IgnoreExitCodes bool `mapstructure:"ignore_exit_codes"`

	// If true, `sudo` will NOT be used to execute Puppet.
	PreventSudo bool `mapstructure:"prevent_sudo"`

	// The directory that contains the puppet binary.
	// E.g. if it can't be found on the standard path.
	PuppetBinDir string `mapstructure:"puppet_bin_dir"`

	// The hostname of the Puppet node.
	PuppetNode string `mapstructure:"puppet_node"`

	// The hostname of the Puppet server.
	PuppetServer string `mapstructure:"puppet_server"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_dir"`

	// The directory from which the command will be executed.
	// Packer requires the directory to exist when running puppet.
	WorkingDir string `mapstructure:"working_directory"`

	// Instructs the communicator to run the remote script as a Windows
	// scheduled task, effectively elevating the remote user by impersonating
	// a logged-in user
	ElevatedUser     string `mapstructure:"elevated_user"`
	ElevatedPassword string `mapstructure:"elevated_password"`
}

type guestOSTypeConfig struct {
	executeCommand   string
	facterVarsFmt    string
	facterVarsJoiner string
	stagingDir       string
	tempDir          string
}

// FIXME assumes both Packer host and target are same OS
var guestOSTypeConfigs = map[string]guestOSTypeConfig{
	provisioner.UnixOSType: {
		tempDir:    "/tmp",
		stagingDir: "/tmp/packer-puppet-server",
		executeCommand: "cd {{.WorkingDir}} && " +
			`{{if ne .FacterVars ""}}{{.FacterVars}} {{end}}` +
			"{{if .Sudo}}sudo -E {{end}}" +
			`{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}` +
			"puppet agent --onetime --no-daemonize --detailed-exitcodes " +
			"{{if .Debug}}--debug {{end}}" +
			`{{if ne .PuppetServer ""}}--server='{{.PuppetServer}}' {{end}}` +
			`{{if ne .PuppetNode ""}}--certname='{{.PuppetNode}}' {{end}}` +
			`{{if ne .ClientCertPath ""}}--certdir='{{.ClientCertPath}}' {{end}}` +
			`{{if ne .ClientPrivateKeyPath ""}}--privatekeydir='{{.ClientPrivateKeyPath}}' {{end}}` +
			`{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}`,
		facterVarsFmt:    "FACTER_%s='%s'",
		facterVarsJoiner: " ",
	},
	provisioner.WindowsOSType: {
		tempDir:    filepath.ToSlash(os.Getenv("TEMP")),
		stagingDir: filepath.ToSlash(os.Getenv("SYSTEMROOT")) + "/Temp/packer-puppet-server",
		executeCommand: "cd {{.WorkingDir}} && " +
			`{{if ne .FacterVars ""}}{{.FacterVars}} && {{end}}` +
			`{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}` +
			"puppet agent --onetime --no-daemonize --detailed-exitcodes " +
			"{{if .Debug}}--debug {{end}}" +
			`{{if ne .PuppetServer ""}}--server='{{.PuppetServer}}' {{end}}` +
			`{{if ne .PuppetNode ""}}--certname='{{.PuppetNode}}' {{end}}` +
			`{{if ne .ClientCertPath ""}}--certdir='{{.ClientCertPath}}' {{end}}` +
			`{{if ne .ClientPrivateKeyPath ""}}--privatekeydir='{{.ClientPrivateKeyPath}}' {{end}}` +
			`{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}`,
		facterVarsFmt:    `SET "FACTER_%s=%s"`,
		facterVarsJoiner: " & ",
	},
}

type Provisioner struct {
	config            Config
	communicator      packer.Communicator
	guestOSTypeConfig guestOSTypeConfig
	guestCommands     *provisioner.GuestCommands
}

type ExecuteTemplate struct {
	ClientCertPath       string
	ClientPrivateKeyPath string
	Debug                bool
	ExtraArguments       string
	FacterVars           string
	PuppetNode           string
	PuppetServer         string
	PuppetBinDir         string
	Sudo                 bool
	WorkingDir           string
}

type EnvVarsTemplate struct {
	WinRMPassword string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	// Create passthrough for winrm password so we can fill it in once we know
	// it
	p.config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: `{{.WinRMPassword}}`,
	}

	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
				"extra_arguments",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.GuestOSType == "" {
		p.config.GuestOSType = provisioner.DefaultOSType
	}
	p.config.GuestOSType = strings.ToLower(p.config.GuestOSType)

	var ok bool
	p.guestOSTypeConfig, ok = guestOSTypeConfigs[p.config.GuestOSType]
	if !ok {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	p.guestCommands, err = provisioner.NewGuestCommands(p.config.GuestOSType, !p.config.PreventSudo)
	if err != nil {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = p.guestOSTypeConfig.executeCommand
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = p.guestOSTypeConfig.stagingDir
	}

	if p.config.WorkingDir == "" {
		p.config.WorkingDir = p.config.StagingDir
	}

	if p.config.Facter == nil {
		p.config.Facter = make(map[string]string)
	}
	p.config.Facter["packer_build_name"] = p.config.PackerBuildName
	p.config.Facter["packer_builder_type"] = p.config.PackerBuilderType

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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Puppet...")
	p.communicator = comm
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
		facterVars = append(facterVars, fmt.Sprintf(p.guestOSTypeConfig.facterVarsFmt, k, v))
	}

	data := ExecuteTemplate{
		ClientCertPath:       remoteClientCertPath,
		ClientPrivateKeyPath: remoteClientPrivateKeyPath,
		ExtraArguments:       "",
		FacterVars:           strings.Join(facterVars, p.guestOSTypeConfig.facterVarsJoiner),
		PuppetNode:           p.config.PuppetNode,
		PuppetServer:         p.config.PuppetServer,
		PuppetBinDir:         p.config.PuppetBinDir,
		Sudo:                 !p.config.PreventSudo,
		WorkingDir:           p.config.WorkingDir,
	}

	p.config.ctx.Data = &data
	_ExtraArguments, err := interpolate.Render(strings.Join(p.config.ExtraArguments, " "), &p.config.ctx)
	if err != nil {
		return err
	}
	data.ExtraArguments = _ExtraArguments

	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	if p.config.ElevatedUser != "" {
		command, err = provisioner.GenerateElevatedRunner(command, p)
		if err != nil {
			return err
		}
	}

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	ui.Message(fmt.Sprintf("Running Puppet: %s", command))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 && cmd.ExitStatus() != 2 && !p.config.IgnoreExitCodes {
		return fmt.Errorf("Puppet exited with a non-zero exit status: %d", cmd.ExitStatus())
	}

	if p.config.CleanStagingDir {
		if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error removing staging directory: %s", err)
		}
	}

	return nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	ctx := context.TODO()

	cmd := &packer.RemoteCmd{Command: p.guestCommands.CreateDir(dir)}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	// Chmod the directory to 0777 just so that we can access it as our user
	cmd = &packer.RemoteCmd{Command: p.guestCommands.Chmod(dir, "0777")}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ctx := context.TODO()

	cmd := &packer.RemoteCmd{Command: p.guestCommands.RemoveDir(dir)}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
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

func getWinRMPassword(buildName string) string {
	winRMPass, _ := commonhelper.RetrieveSharedState("winrm_password", buildName)
	packer.LogSecretFilter.Set(winRMPass)
	return winRMPass
}

func (p *Provisioner) Communicator() packer.Communicator {
	return p.communicator
}

func (p *Provisioner) ElevatedUser() string {
	return p.config.ElevatedUser
}

func (p *Provisioner) ElevatedPassword() string {
	// Replace ElevatedPassword for winrm users who used this feature
	p.config.ctx.Data = &EnvVarsTemplate{
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}

	elevatedPassword, _ := interpolate.Render(p.config.ElevatedPassword, &p.config.ctx)

	return elevatedPassword
}
