// This package implements a provisioner for Packer that executes
// Puppet on the remote machine, configured to apply a local manifest
// versus connecting to a Puppet master.
package puppetmasterless

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	tpl                 *packer.ConfigTemplate

	// The command used to execute Puppet.
	ExecuteCommand string `mapstructure:"execute_command"`

	// Additional facts to set when executing Puppet
	Facter map[string]string

	// Path to a hiera configuration file to upload and use.
	HieraConfigPath string `mapstructure:"hiera_config_path"`

	// An array of local paths of modules to upload.
	ModulePaths []string `mapstructure:"module_paths"`

	// The main manifest file to apply to kick off the entire thing.
	ManifestFile string `mapstructure:"manifest_file"`

	// A directory of manifest files that will be uploaded to the remote
	// machine.
	ManifestDir string `mapstructure:"manifest_dir"`

	// If true, `sudo` will NOT be used to execute Puppet.
	PreventSudo bool `mapstructure:"prevent_sudo"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_directory"`
}

type Provisioner struct {
	config Config
}

type ExecuteTemplate struct {
	FacterVars      string
	HieraConfigPath string
	ModulePath      string
	ManifestFile    string
	ManifestDir     string
	Sudo            bool
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

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	// Set some defaults
	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "{{.FacterVars}} {{if .Sudo}} sudo -E {{end}}" +
			"puppet apply --verbose --modulepath='{{.ModulePath}}' " +
			"{{if ne .HieraConfigPath \"\"}}--hiera_config='{{.HieraConfigPath}}' {{end}}" +
			"{{if ne .ManifestDir \"\"}}--manifestdir='{{.ManifestDir}}' {{end}}" +
			"--detailed-exitcodes " +
			"{{.ManifestFile}}"
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-puppet-masterless"
	}

	// Templates
	templates := map[string]*string{
		"hiera_config_path": &p.config.HieraConfigPath,
		"manifest_file":     &p.config.ManifestFile,
		"manifest_dir":      &p.config.ManifestDir,
		"staging_dir":       &p.config.StagingDir,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	sliceTemplates := map[string][]string{
		"module_paths": p.config.ModulePaths,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = p.config.tpl.Process(elem, nil)
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	validates := map[string]*string{
		"execute_command": &p.config.ExecuteCommand,
	}

	for n, ptr := range validates {
		if err := p.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
		}
	}

	newFacts := make(map[string]string)
	for k, v := range p.config.Facter {
		k, err := p.config.tpl.Process(k, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing facter key %s: %s", k, err))
			continue
		}

		v, err := p.config.tpl.Process(v, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing facter value '%s': %s", v, err))
			continue
		}

		newFacts[k] = v
	}

	p.config.Facter = newFacts

	// Validation
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
		info, err := os.Stat(p.config.ManifestFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("manifest_file is invalid: %s", err))
		} else if info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("manifest_file must point to a file"))
		}
	}

	for i, path := range p.config.ModulePaths {
		info, err := os.Stat(path)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("module_path[%d] is invalid: %s", i, err))
		} else if !info.IsDir() {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("module_path[%d] must point to a directory"))
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
		facterVars = append(facterVars, fmt.Sprintf("FACTER_%s='%s'", k, v))
	}

	// Execute Puppet
	command, err := p.config.tpl.Process(p.config.ExecuteCommand, &ExecuteTemplate{
		FacterVars:      strings.Join(facterVars, " "),
		HieraConfigPath: remoteHieraConfigPath,
		ManifestDir:     remoteManifestDir,
		ManifestFile:    remoteManifestFile,
		ModulePath:      strings.Join(modulePaths, ":"),
		Sudo:            !p.config.PreventSudo,
	})
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

	if cmd.ExitStatus != 0 && cmd.ExitStatus != 2 {
		return fmt.Errorf("Puppet exited with a non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}

func (p *Provisioner) uploadHieraConfig(ui packer.Ui, comm packer.Communicator) (string, error) {
	ui.Message("Uploading hiera configuration...")
	f, err := os.Open(p.config.HieraConfigPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	path := fmt.Sprintf("%s/hiera.yaml", p.config.StagingDir)
	if err := comm.Upload(path, f); err != nil {
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

	// Upload the main manifest
	f, err := os.Open(p.config.ManifestFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	manifestFilename := filepath.Base(p.config.ManifestFile)
	remoteManifestFile := fmt.Sprintf("%s/%s", remoteManifestsPath, manifestFilename)
	if err := comm.Upload(remoteManifestFile, f); err != nil {
		return "", err
	}

	return remoteManifestFile, nil
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
