package vagrant

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Builder implements packer.Builder and builds the actual VirtualBox
// images.
type Builder struct {
	config *Config
	runner multistep.Runner
}

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
}

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	common.HTTPConfig      `mapstructure:",squash"`
	common.ISOConfig       `mapstructure:",squash"`
	common.FloppyConfig    `mapstructure:",squash"`
	bootcommand.BootConfig `mapstructure:",squash"`
	SSHConfig              `mapstructure:",squash"`

	// This is the name of the new virtual machine.
	// By default this is "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	OutputDir    string `mapstructure:"output_dir"`
	SourceBox    string `mapstructure:"source_path"`
	GlobalID     string `mapstructure:"global_id"`
	Checksum     string `mapstructure:"checksum"`
	ChecksumType string `mapstructure:"checksum_type"`
	BoxName      string `mapstructure:"box_name"`

	Provider string `mapstructure:"provider"`

	Communicator string `mapstructure:"communicator"`

	// What vagrantfile to use
	VagrantfileTpl string `mapstructure:"vagrantfile_template"`

	// Whether to Halt, Suspend, or Destroy the box
	TeardownMethod string `mapstructure:"teardown_method"`

	// Options for the "vagrant init" command
	BoxVersion   string `mapstructure:"box_version"`
	Template     string `mapstructure:"template"`
	SyncedFolder string `mapstructure:"synced_folder"`

	// Options for the "vagrant box add" command
	SkipAdd     bool   `mapstructure:"skip_add"`
	AddCACert   string `mapstructure:"add_cacert"`
	AddCAPath   string `mapstructure:"add_capath"`
	AddCert     string `mapstructure:"add_cert"`
	AddClean    bool   `mapstructure:"add_clean"`
	AddForce    bool   `mapstructure:"add_force"`
	AddInsecure bool   `mapstructure:"add_insecure"`

	// Don't package the Vagrant box after build.
	SkipPackage       bool     `mapstructure:"skip_package"`
	OutputVagrantfile string   `mapstructure:"output_vagrantfile"`
	PackageInclude    []string `mapstructure:"package_include"`

	ctx interpolate.Context
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config = new(Config)
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.Comm.SSHTimeout == 0 {
		b.config.Comm.SSHTimeout = 10 * time.Minute
	}

	if b.config.Comm.Type != "ssh" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf(`The Vagrant builder currently only supports the ssh communicator"`))
	}
	// The box isn't a namespace like you'd pull from vagrant cloud
	if b.config.BoxName == "" {
		b.config.BoxName = fmt.Sprintf("packer_%s", b.config.PackerBuildName)
	}

	if b.config.SourceBox == "" {
		if b.config.GlobalID == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required unless you have set global_id"))
		}
	} else {
		if b.config.GlobalID != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("You may either set global_id or source_path but not both"))
		}
		if strings.HasSuffix(b.config.SourceBox, ".box") {
			b.config.SourceBox, err = common.ValidatedURL(b.config.SourceBox)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is invalid: %s", err))
			}
			fileOK := common.FileExistsLocally(b.config.SourceBox)
			if !fileOK {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Source file '%s' needs to exist at time of config validation!", b.config.SourceBox))
			}
		}
	}

	if b.config.TeardownMethod == "" {
		// If we're using a box that's already opened on the system, don't
		// automatically destroy it. If we open the box ourselves, then go ahead
		// and kill it by default.
		if b.config.GlobalID != "" {
			b.config.TeardownMethod = "halt"
		} else {
			b.config.TeardownMethod = "destroy"
		}
	} else {
		matches := false
		for _, name := range []string{"halt", "suspend", "destroy"} {
			if strings.ToLower(b.config.TeardownMethod) == name {
				matches = true
			}
		}
		if !matches {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf(`TeardownMethod must be "halt", "suspend", or "destroy"`))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a VirtualBox appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with VirtualBox
	VagrantCWD, err := filepath.Abs(b.config.OutputDir)
	if err != nil {
		return nil, err
	}
	driver, err := NewDriver(VagrantCWD)
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("cache", cache)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{}
	// Download if source box isn't from vagrant cloud.
	if strings.HasSuffix(b.config.SourceBox, ".box") {
		steps = append(steps, &common.StepDownload{
			Checksum:     b.config.Checksum,
			ChecksumType: b.config.ChecksumType,
			Description:  "Box",
			Extension:    "box",
			ResultKey:    "box_path",
			Url:          []string{b.config.SourceBox},
		})
	}
	steps = append(steps,
		&common.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&StepCreateVagrantfile{
			Template:     b.config.Template,
			SyncedFolder: b.config.SyncedFolder,
			SourceBox:    b.config.SourceBox,
			OutputDir:    b.config.OutputDir,
			GlobalID:     b.config.GlobalID,
		},
		&StepAddBox{
			BoxVersion:   b.config.BoxVersion,
			CACert:       b.config.AddCACert,
			CAPath:       b.config.AddCAPath,
			DownloadCert: b.config.AddCert,
			Clean:        b.config.AddClean,
			Force:        b.config.AddForce,
			Insecure:     b.config.AddInsecure,
			Provider:     b.config.Provider,
			SourceBox:    b.config.SourceBox,
			BoxName:      b.config.BoxName,
			GlobalID:     b.config.GlobalID,
			SkipAdd:      b.config.SkipAdd,
		},
		&StepUp{
			TeardownMethod: b.config.TeardownMethod,
			Provider:       b.config.Provider,
			GlobalID:       b.config.GlobalID,
		},
		&StepSSHConfig{
			b.config.GlobalID,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      CommHost(),
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
		new(common.StepProvision),
		&StepPackage{
			SkipPackage: b.config.SkipPackage,
			Include:     b.config.PackageInclude,
			Vagrantfile: b.config.OutputVagrantfile,
			GlobalID:    b.config.GlobalID,
		})

	// Run the steps.
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
