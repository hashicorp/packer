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

	// An array of local paths of modules to upload.
	ModulePaths []string `mapstructure:"module_paths"`

	// The main manifest file to apply to kick off the entire thing.
	ManifestFile string `mapstructure:"manifest_file"`

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
	ModulePath   string
	ManifestFile string
	Sudo         bool
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
		p.config.ExecuteCommand = "{{if .Sudo}}sudo {{end}}puppet apply --verbose --modulepath='{{.ModulePath}}' {{.ManifestFile}}"
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-puppet-masterless"
	}

	// Templates
	templates := map[string]*string{
		"staging_dir": &p.config.StagingDir,
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

	// Validation
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

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Creating Puppet staging directory...")
	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	// Upload all modules
	modulePaths := make([]string, 0, len(p.config.ModulePaths))
	for i, path := range p.config.ModulePaths {
		ui.Message(fmt.Sprintf("Upload local modules from: %s", path))
		targetPath := fmt.Sprintf("%s/module-%d", p.config.StagingDir, i)
		if err := p.uploadDirectory(ui, comm, targetPath, path); err != nil {
			return fmt.Errorf("Error uploading modules: %s", err)
		}

		modulePaths = append(modulePaths, targetPath)
	}

	// Upload manifests
	ui.Message("Uploading manifests...")
	remoteManifestFile, err := p.uploadManifests(ui, comm)
	if err != nil {
		return fmt.Errorf("Error uploading manifests: %s", err)
	}

	// Execute Puppet
	command, err := p.config.tpl.Process(p.config.ExecuteCommand, &ExecuteTemplate{
		ManifestFile: remoteManifestFile,
		ModulePath:   strings.Join(modulePaths, ":"),
		Sudo:         !p.config.PreventSudo,
	})
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	ui.Message("Running Puppet...")
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Puppet exited with a non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
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
