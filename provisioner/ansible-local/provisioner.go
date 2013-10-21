package ansiblelocal

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
	"path/filepath"
)

const DefaultStagingDir = "/tmp/packer-provisioner-ansible-local"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	tpl                 *packer.ConfigTemplate

	// The main playbook file to execute.
	PlaybookFile string `mapstructure:"playbook_file"`

	// An array of local paths of playbook files to upload.
	PlaybookPaths []string `mapstructure:"playbook_paths"`

	// An array of local paths of roles to upload.
	RolePaths []string `mapstructure:"role_paths"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_directory"`
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

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if p.config.StagingDir == "" {
		p.config.StagingDir = DefaultStagingDir
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

	// Validation
	err = validateFileConfig(p.config.PlaybookFile, "playbook_file", true)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}
	for _, path := range p.config.PlaybookPaths {
		err := validateFileConfig(path, "playbook_paths", false)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}
	for _, path := range p.config.RolePaths {
		if err := validateDirConfig(path, "role_paths"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}
	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Ansible...")

	ui.Message("Creating Ansible staging directory...")
	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	ui.Message("Uploading main Playbook file...")
	src := p.config.PlaybookFile
	dst := filepath.Join(p.config.StagingDir, filepath.Base(src))
	if err := p.uploadFile(ui, comm, dst, src); err != nil {
		return fmt.Errorf("Error uploading main playbook: %s", err)
	}

	if len(p.config.RolePaths) > 0 {
		ui.Message("Uploading role directories...")
		for _, src := range p.config.RolePaths {
			dst := filepath.Join(p.config.StagingDir, "roles", filepath.Base(src))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading roles: %s", err)
			}
		}
	}

	if len(p.config.PlaybookPaths) > 0 {
		ui.Message("Uploading additional Playbooks...")
		if err := p.createDir(ui, comm, filepath.Join(p.config.StagingDir, "playbooks")); err != nil {
			return fmt.Errorf("Error creating playbooks directory: %s", err)
		}
		for _, src := range p.config.PlaybookPaths {
			dst := filepath.Join(p.config.StagingDir, "playbooks", filepath.Base(src))
			if err := p.uploadFile(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading playbooks: %s", err)
			}
		}
	}

	if err := p.executeAnsible(ui, comm); err != nil {
		return fmt.Errorf("Error executing Ansible: %s", err)
	}
	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}

func (p *Provisioner) executeAnsible(ui packer.Ui, comm packer.Communicator) error {
	playbook := filepath.Join(p.config.StagingDir, filepath.Base(p.config.PlaybookFile))

	// The inventory must be set to "127.0.0.1,".  The comma is important
	// as its the only way to override the ansible inventory when dealing
	// with a single host.
	command := fmt.Sprintf("ansible-playbook %s -c local -i %s", playbook, `"127.0.0.1,"`)

	ui.Message(fmt.Sprintf("Executing Ansible: %s", command))
	cmd := &packer.RemoteCmd{
		Command: command,
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus)
	}
	return nil
}

func validateDirConfig(path string, config string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s: %s must point to a directory", config, path)
	}
	return nil
}

func validateFileConfig(name string, config string, req bool) error {
	if req {
		if name == "" {
			return fmt.Errorf("%s must be specified.", config)
		}
	}
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, name, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, name)
	}
	return nil
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening: %s", err)
	}
	defer f.Close()

	if err = comm.Upload(dst, f); err != nil {
		return fmt.Errorf("Error uploading %s: %s", src, err)
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

func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string) error {
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
