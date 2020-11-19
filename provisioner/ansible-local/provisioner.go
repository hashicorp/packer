//go:generate mapstructure-to-hcl2 -type Config

package ansiblelocal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

const DefaultStagingDir = "/tmp/packer-provisioner-ansible-local"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// The command to run ansible
	Command string

	// Extra options to pass to the ansible command
	ExtraArguments []string `mapstructure:"extra_arguments"`

	// Path to group_vars directory
	GroupVars string `mapstructure:"group_vars"`

	// Path to host_vars directory
	HostVars string `mapstructure:"host_vars"`

	// The playbook dir to upload.
	PlaybookDir string `mapstructure:"playbook_dir"`

	// The main playbook file to execute.
	PlaybookFile string `mapstructure:"playbook_file"`

	// The playbook files to execute.
	PlaybookFiles []string `mapstructure:"playbook_files"`

	// An array of local paths of playbook files to upload.
	PlaybookPaths []string `mapstructure:"playbook_paths"`

	// An array of local paths of roles to upload.
	RolePaths []string `mapstructure:"role_paths"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_directory"`

	// If true, staging directory is removed after executing ansible.
	CleanStagingDir bool `mapstructure:"clean_staging_directory"`

	// The optional inventory file
	InventoryFile string `mapstructure:"inventory_file"`

	// The optional inventory groups
	InventoryGroups []string `mapstructure:"inventory_groups"`

	// The optional ansible-galaxy requirements file
	GalaxyFile string `mapstructure:"galaxy_file"`

	// The command to run ansible-galaxy
	GalaxyCommand string `mapstructure:"galaxy_command"`
}

type Provisioner struct {
	config Config

	playbookFiles []string
	generatedData map[string]interface{}
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "ansible-local",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Reset the state.
	p.playbookFiles = make([]string, 0, len(p.config.PlaybookFiles))

	// Defaults
	if p.config.Command == "" {
		p.config.Command = "ANSIBLE_FORCE_COLOR=1 PYTHONUNBUFFERED=1 ansible-playbook"
	}
	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = filepath.ToSlash(filepath.Join(DefaultStagingDir, uuid.TimeOrderedUUID()))
	}

	// Validation
	var errs *packersdk.MultiError

	// Check that either playbook_file or playbook_files is specified
	if len(p.config.PlaybookFiles) != 0 && p.config.PlaybookFile != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Either playbook_file or playbook_files can be specified, not both"))
	}
	if len(p.config.PlaybookFiles) == 0 && p.config.PlaybookFile == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Either playbook_file or playbook_files must be specified"))
	}
	if p.config.PlaybookFile != "" {
		err = validateFileConfig(p.config.PlaybookFile, "playbook_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for _, playbookFile := range p.config.PlaybookFiles {
		if err := validateFileConfig(playbookFile, "playbook_files", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		} else {
			playbookFile, err := filepath.Abs(playbookFile)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
			} else {
				p.playbookFiles = append(p.playbookFiles, playbookFile)
			}
		}
	}

	// Check that the inventory file exists, if configured
	if len(p.config.InventoryFile) > 0 {
		err = validateFileConfig(p.config.InventoryFile, "inventory_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the galaxy file exists, if configured
	if len(p.config.GalaxyFile) > 0 {
		err = validateFileConfig(p.config.GalaxyFile, "galaxy_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the playbook_dir directory exists, if configured
	if len(p.config.PlaybookDir) > 0 {
		if err := validateDirConfig(p.config.PlaybookDir, "playbook_dir"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the group_vars directory exists, if configured
	if len(p.config.GroupVars) > 0 {
		if err := validateDirConfig(p.config.GroupVars, "group_vars"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the host_vars directory exists, if configured
	if len(p.config.HostVars) > 0 {
		if err := validateDirConfig(p.config.HostVars, "host_vars"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for _, path := range p.config.PlaybookPaths {
		err := validateDirConfig(path, "playbook_paths")
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}
	for _, path := range p.config.RolePaths {
		if err := validateDirConfig(path, "role_paths"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible...")
	p.generatedData = generatedData

	if len(p.config.PlaybookDir) > 0 {
		ui.Message("Uploading Playbook directory to Ansible staging directory...")
		if err := p.uploadDir(ui, comm, p.config.StagingDir, p.config.PlaybookDir); err != nil {
			return fmt.Errorf("Error uploading playbook_dir directory: %s", err)
		}
	} else {
		ui.Message("Creating Ansible staging directory...")
		if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error creating staging directory: %s", err)
		}
	}

	if p.config.PlaybookFile != "" {
		ui.Message("Uploading main Playbook file...")
		src := p.config.PlaybookFile
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
		if err := p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading main playbook: %s", err)
		}
	} else if err := p.provisionPlaybookFiles(ui, comm); err != nil {
		return err
	}

	if len(p.config.InventoryFile) == 0 {
		tf, err := tmp.File("packer-provisioner-ansible-local")
		if err != nil {
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		defer os.Remove(tf.Name())
		if len(p.config.InventoryGroups) != 0 {
			content := ""
			for _, group := range p.config.InventoryGroups {
				content += fmt.Sprintf("[%s]\n127.0.0.1\n", group)
			}
			_, err = tf.Write([]byte(content))
		} else {
			_, err = tf.Write([]byte("127.0.0.1"))
		}
		if err != nil {
			tf.Close()
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		tf.Close()
		p.config.InventoryFile = tf.Name()
		defer func() {
			p.config.InventoryFile = ""
		}()
	}

	if len(p.config.GalaxyFile) > 0 {
		ui.Message("Uploading galaxy file...")
		src := p.config.GalaxyFile
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
		if err := p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading galaxy file: %s", err)
		}
	}

	ui.Message("Uploading inventory file...")
	src := p.config.InventoryFile
	dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
	if err := p.uploadFile(ui, comm, dst, src); err != nil {
		return fmt.Errorf("Error uploading inventory file: %s", err)
	}

	if len(p.config.GroupVars) > 0 {
		ui.Message("Uploading group_vars directory...")
		src := p.config.GroupVars
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "group_vars"))
		if err := p.uploadDir(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading group_vars directory: %s", err)
		}
	}

	if len(p.config.HostVars) > 0 {
		ui.Message("Uploading host_vars directory...")
		src := p.config.HostVars
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "host_vars"))
		if err := p.uploadDir(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading host_vars directory: %s", err)
		}
	}

	if len(p.config.RolePaths) > 0 {
		ui.Message("Uploading role directories...")
		for _, src := range p.config.RolePaths {
			dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "roles", filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading roles: %s", err)
			}
		}
	}

	if len(p.config.PlaybookPaths) > 0 {
		ui.Message("Uploading additional Playbooks...")
		playbookDir := filepath.ToSlash(filepath.Join(p.config.StagingDir, "playbooks"))
		if err := p.createDir(ui, comm, playbookDir); err != nil {
			return fmt.Errorf("Error creating playbooks directory: %s", err)
		}
		for _, src := range p.config.PlaybookPaths {
			dst := filepath.ToSlash(filepath.Join(playbookDir, filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading playbooks: %s", err)
			}
		}
	}

	if err := p.executeAnsible(ui, comm); err != nil {
		return fmt.Errorf("Error executing Ansible: %s", err)
	}

	if p.config.CleanStagingDir {
		ui.Message("Removing staging directory...")
		if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error removing staging directory: %s", err)
		}
	}
	return nil
}

func (p *Provisioner) provisionPlaybookFiles(ui packersdk.Ui, comm packer.Communicator) error {
	var playbookDir string
	if p.config.PlaybookDir != "" {
		var err error
		playbookDir, err = filepath.Abs(p.config.PlaybookDir)
		if err != nil {
			return err
		}
	}
	for index, playbookFile := range p.playbookFiles {
		if playbookDir != "" && strings.HasPrefix(playbookFile, playbookDir) {
			p.playbookFiles[index] = strings.TrimPrefix(playbookFile, playbookDir)
			continue
		}
		if err := p.provisionPlaybookFile(ui, comm, playbookFile); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provisioner) provisionPlaybookFile(ui packersdk.Ui, comm packer.Communicator, playbookFile string) error {
	ui.Message(fmt.Sprintf("Uploading playbook file: %s", playbookFile))

	remoteDir := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Dir(playbookFile)))
	remotePlaybookFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, playbookFile))

	if err := p.createDir(ui, comm, remoteDir); err != nil {
		return fmt.Errorf("Error uploading playbook file: %s [%s]", playbookFile, err)
	}

	if err := p.uploadFile(ui, comm, remotePlaybookFile, playbookFile); err != nil {
		return fmt.Errorf("Error uploading playbook: %s [%s]", playbookFile, err)
	}

	return nil
}

func (p *Provisioner) executeGalaxy(ui packersdk.Ui, comm packer.Communicator) error {
	ctx := context.TODO()
	rolesDir := filepath.ToSlash(filepath.Join(p.config.StagingDir, "roles"))
	galaxyFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.GalaxyFile)))

	// ansible-galaxy install -r requirements.yml -p roles/
	command := fmt.Sprintf("cd %s && %s install -r %s -p %s",
		p.config.StagingDir, p.config.GalaxyCommand, galaxyFile, rolesDir)
	ui.Message(fmt.Sprintf("Executing Ansible Galaxy: %s", command))
	cmd := &packer.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		// ansible-galaxy version 2.0.0.2 doesn't return exit codes on error..
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
	}
	return nil
}

func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packer.Communicator) error {
	inventory := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.InventoryFile)))

	extraArgs := fmt.Sprintf(" --extra-vars \"packer_build_name=%s packer_builder_type=%s packer_http_addr=%s -o IdentitiesOnly=yes\" ",
		p.config.PackerBuildName, p.config.PackerBuilderType, p.generatedData["PackerHTTPAddr"])
	if len(p.config.ExtraArguments) > 0 {
		extraArgs = extraArgs + strings.Join(p.config.ExtraArguments, " ")
	}

	// Fetch external dependencies
	if len(p.config.GalaxyFile) > 0 {
		if err := p.executeGalaxy(ui, comm); err != nil {
			return fmt.Errorf("Error executing Ansible Galaxy: %s", err)
		}
	}

	if p.config.PlaybookFile != "" {
		playbookFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.PlaybookFile)))
		if err := p.executeAnsiblePlaybook(ui, comm, playbookFile, extraArgs, inventory); err != nil {
			return err
		}
	}

	for _, playbookFile := range p.playbookFiles {
		playbookFile = filepath.ToSlash(filepath.Join(p.config.StagingDir, playbookFile))
		if err := p.executeAnsiblePlaybook(ui, comm, playbookFile, extraArgs, inventory); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provisioner) executeAnsiblePlaybook(
	ui packersdk.Ui, comm packer.Communicator, playbookFile, extraArgs, inventory string,
) error {
	ctx := context.TODO()
	command := fmt.Sprintf("cd %s && %s %s%s -c local -i %s",
		p.config.StagingDir, p.config.Command, playbookFile, extraArgs, inventory,
	)
	ui.Message(fmt.Sprintf("Executing Ansible: %s", command))
	cmd := &packer.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		if cmd.ExitStatus() == 127 {
			return fmt.Errorf("%s could not be found. Verify that it is available on the\n"+
				"PATH after connecting to the machine.",
				p.config.Command)
		}

		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
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

func (p *Provisioner) uploadFile(ui packersdk.Ui, comm packer.Communicator, dst, src string) error {
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

func (p *Provisioner) createDir(ui packersdk.Ui, comm packer.Communicator, dir string) error {
	ctx := context.TODO()
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}

	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packersdk.Ui, comm packer.Communicator, dir string) error {
	ctx := context.TODO()
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("rm -rf '%s'", dir),
	}

	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) uploadDir(ui packersdk.Ui, comm packer.Communicator, dst, src string) error {
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
