//go:generate mapstructure-to-hcl2 -type Config

// Package puppetmasterless implements a provisioner for Packer that executes
// Puppet on the remote machine, configured to apply a local manifest
// versus connecting to a Puppet master.
package puppetmasterless

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

	// The Guest OS Type (unix or windows)
	GuestOSType string `mapstructure:"guest_os_type"`

	// The command used to execute Puppet.
	ExecuteCommand string `mapstructure:"execute_command"`

	// Additional arguments to pass when executing Puppet
	ExtraArguments []string `mapstructure:"extra_arguments"`

	// Additional facts to set when executing Puppet
	Facter map[string]string

	// Path to a hiera configuration file to upload and use.
	HieraConfigPath string `mapstructure:"hiera_config_path"`

	// If true, packer will ignore all exit-codes from a puppet run
	IgnoreExitCodes bool `mapstructure:"ignore_exit_codes"`

	// An array of local paths of modules to upload.
	ModulePaths []string `mapstructure:"module_paths"`

	// The main manifest file to apply to kick off the entire thing.
	ManifestFile string `mapstructure:"manifest_file"`

	// A directory of manifest files that will be uploaded to the remote
	// machine.
	ManifestDir string `mapstructure:"manifest_dir"`

	// If true, `sudo` will NOT be used to execute Puppet.
	PreventSudo bool `mapstructure:"prevent_sudo"`

	// The directory that contains the puppet binary.
	// E.g. if it can't be found on the standard path.
	PuppetBinDir string `mapstructure:"puppet_bin_dir"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_directory"`

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
	modulePathJoiner string
	stagingDir       string
	tempDir          string
}

// FIXME assumes both Packer host and target are same OS
var guestOSTypeConfigs = map[string]guestOSTypeConfig{
	provisioner.UnixOSType: {
		tempDir:    "/tmp",
		stagingDir: "/tmp/packer-puppet-masterless",
		executeCommand: "cd {{.WorkingDir}} && " +
			`{{if ne .FacterVars ""}}{{.FacterVars}} {{end}}` +
			"{{if .Sudo}}sudo -E {{end}}" +
			`{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}` +
			"puppet apply --detailed-exitcodes " +
			"{{if .Debug}}--debug {{end}}" +
			`{{if ne .ModulePath ""}}--modulepath='{{.ModulePath}}' {{end}}` +
			`{{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}}` +
			`{{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}}` +
			`{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}` +
			"{{.ManifestFile}}",
		facterVarsFmt:    "FACTER_%s='%s'",
		facterVarsJoiner: " ",
		modulePathJoiner: ":",
	},
	provisioner.WindowsOSType: {
		tempDir:    filepath.ToSlash(os.Getenv("TEMP")),
		stagingDir: filepath.ToSlash(os.Getenv("SYSTEMROOT")) + "/Temp/packer-puppet-masterless",
		executeCommand: "cd {{.WorkingDir}} && " +
			`{{if ne .FacterVars ""}}{{.FacterVars}} && {{end}}` +
			`{{if ne .PuppetBinDir ""}}{{.PuppetBinDir}}/{{end}}` +
			"puppet apply --detailed-exitcodes " +
			"{{if .Debug}}--debug {{end}}" +
			`{{if ne .ModulePath ""}}--modulepath='{{.ModulePath}}' {{end}}` +
			`{{if ne .HieraConfigPath ""}}--hiera_config='{{.HieraConfigPath}}' {{end}}` +
			`{{if ne .ManifestDir ""}}--manifestdir='{{.ManifestDir}}' {{end}}` +
			`{{if ne .ExtraArguments ""}}{{.ExtraArguments}} {{end}}` +
			"{{.ManifestFile}}",
		facterVarsFmt:    `SET "FACTER_%s=%s"`,
		facterVarsJoiner: " & ",
		modulePathJoiner: ";",
	},
}

type Provisioner struct {
	config            Config
	communicator      packer.Communicator
	guestOSTypeConfig guestOSTypeConfig
	guestCommands     *provisioner.GuestCommands
}

type ExecuteTemplate struct {
	Debug            bool
	ExtraArguments   string
	FacterVars       string
	HieraConfigPath  string
	ModulePath       string
	ModulePathJoiner string
	ManifestFile     string
	ManifestDir      string
	PuppetBinDir     string
	Sudo             bool
	WorkingDir       string
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

	// Set some defaults
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

	// Validation
	var errs *packer.MultiError
	if p.config.HieraConfigPath != "" {
		info, err := os.Stat(p.config.HieraConfigPath)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("hiera_config_path is invalid: %s", err))
		} else if info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("hiera_config_path must point to a file"))
		}
	}

	if p.config.ManifestDir != "" {
		info, err := os.Stat(p.config.ManifestDir)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("manifest_dir is invalid: %s", err))
		} else if !info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("manifest_dir must point to a directory"))
		}
	}

	if p.config.ManifestFile == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("A manifest_file must be specified."))
	} else {
		_, err := os.Stat(p.config.ManifestFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("manifest_file is invalid: %s", err))
		}
	}

	for i, path := range p.config.ModulePaths {
		info, err := os.Stat(path)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("module_path[%d] is invalid: %s", i, err))
		} else if !info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("module_path[%d] must point to a directory", i))
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

	// Upload hiera config if set
	remoteHieraConfigPath := ""
	if p.config.HieraConfigPath != "" {
		var err error
		remoteHieraConfigPath, err = p.uploadHieraConfig(ui, comm)
		if err != nil {
			return fmt.Errorf("Error uploading hiera config: %s", err)
		}
	}

	// Upload manifest dir if set
	remoteManifestDir := ""
	if p.config.ManifestDir != "" {
		ui.Message(fmt.Sprintf(
			"Uploading manifest directory from: %s", p.config.ManifestDir))
		remoteManifestDir = fmt.Sprintf("%s/manifests", p.config.StagingDir)
		err := p.uploadDirectory(ui, comm, remoteManifestDir, p.config.ManifestDir)
		if err != nil {
			return fmt.Errorf("Error uploading manifest dir: %s", err)
		}
	}

	// Upload all modules
	modulePaths := make([]string, 0, len(p.config.ModulePaths))
	for i, path := range p.config.ModulePaths {
		ui.Message(fmt.Sprintf("Uploading local modules from: %s", path))
		targetPath := fmt.Sprintf("%s/module-%d", p.config.StagingDir, i)
		if err := p.uploadDirectory(ui, comm, targetPath, path); err != nil {
			return fmt.Errorf("Error uploading modules: %s", err)
		}

		modulePaths = append(modulePaths, targetPath)
	}

	// Upload manifests
	remoteManifestFile, err := p.uploadManifests(ui, comm)
	if err != nil {
		return fmt.Errorf("Error uploading manifests: %s", err)
	}

	// Compile the facter variables
	facterVars := make([]string, 0, len(p.config.Facter))
	for k, v := range p.config.Facter {
		facterVars = append(facterVars, fmt.Sprintf(p.guestOSTypeConfig.facterVarsFmt, k, v))
	}

	data := ExecuteTemplate{
		ExtraArguments:   "",
		FacterVars:       strings.Join(facterVars, p.guestOSTypeConfig.facterVarsJoiner),
		HieraConfigPath:  remoteHieraConfigPath,
		ManifestDir:      remoteManifestDir,
		ManifestFile:     remoteManifestFile,
		ModulePath:       strings.Join(modulePaths, p.guestOSTypeConfig.modulePathJoiner),
		ModulePathJoiner: p.guestOSTypeConfig.modulePathJoiner,
		PuppetBinDir:     p.config.PuppetBinDir,
		Sudo:             !p.config.PreventSudo,
		WorkingDir:       p.config.WorkingDir,
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
		return fmt.Errorf("Got an error starting command: %s", err)
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

func (p *Provisioner) uploadHieraConfig(ui packer.Ui, comm packer.Communicator) (string, error) {
	ui.Message("Uploading hiera configuration...")
	f, err := os.Open(p.config.HieraConfigPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	path := fmt.Sprintf("%s/hiera.yaml", p.config.StagingDir)
	if err := comm.Upload(path, f, nil); err != nil {
		return "", err
	}

	return path, nil
}

func (p *Provisioner) uploadManifests(ui packer.Ui, comm packer.Communicator) (string, error) {
	// Create the remote manifests directory...
	ui.Message("Uploading manifests...")
	remoteManifestsPath := fmt.Sprintf("%s/manifests", p.config.StagingDir)
	if err := p.createDir(ui, comm, remoteManifestsPath); err != nil {
		return "", fmt.Errorf("Error creating manifests directory: %s", err)
	}

	// NOTE! manifest_file may either be a directory or a file, as puppet apply
	// now accepts either one.

	fi, err := os.Stat(p.config.ManifestFile)
	if err != nil {
		return "", fmt.Errorf("Error inspecting manifest file: %s", err)
	}

	if fi.IsDir() {
		// If manifest_file is a directory we'll upload the whole thing
		ui.Message(fmt.Sprintf(
			"Uploading manifest directory from: %s", p.config.ManifestFile))

		remoteManifestDir := fmt.Sprintf("%s/manifests", p.config.StagingDir)
		err := p.uploadDirectory(ui, comm, remoteManifestDir, p.config.ManifestFile)
		if err != nil {
			return "", fmt.Errorf("Error uploading manifest dir: %s", err)
		}
		return remoteManifestDir, nil
	}
	// Otherwise manifest_file is a file and we'll upload it
	ui.Message(fmt.Sprintf(
		"Uploading manifest file from: %s", p.config.ManifestFile))

	f, err := os.Open(p.config.ManifestFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	manifestFilename := filepath.Base(p.config.ManifestFile)
	remoteManifestFile := fmt.Sprintf("%s/%s", remoteManifestsPath, manifestFilename)
	if err := comm.Upload(remoteManifestFile, f, nil); err != nil {
		return "", err
	}
	return remoteManifestFile, nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))

	cmd := &packer.RemoteCmd{Command: p.guestCommands.CreateDir(dir)}
	ctx := context.TODO()

	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status.")
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
