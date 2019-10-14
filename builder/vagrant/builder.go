//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package vagrant

import (
	"context"
	"errors"
	"fmt"
	"os"
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

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	common.HTTPConfig      `mapstructure:",squash"`
	common.ISOConfig       `mapstructure:",squash"`
	common.FloppyConfig    `mapstructure:",squash"`
	bootcommand.BootConfig `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`
	// The directory to create that will contain your output box. We always
	// create this directory and run from inside of it to prevent Vagrant init
	// collisions. If unset, it will be set to packer- plus your buildname.
	OutputDir string `mapstructure:"output_dir" required:"false"`
	// URL of the vagrant box to use, or the name of the vagrant box.
	// hashicorp/precise64, ./mylocalbox.box and https://example.com/my-box.box
	// are all valid source boxes. If your source is a .box file, whether
	// locally or from a URL like the latter example above, you will also need
	// to provide a box_name. This option is required, unless you set
	// global_id. You may only set one or the other, not both.
	SourceBox string `mapstructure:"source_path" required:"true"`
	// the global id of a Vagrant box already added to Vagrant on your system.
	// You can find the global id of your Vagrant boxes using the command
	// vagrant global-status; your global_id will be a 7-digit number and
	// letter comination that you'll find in the leftmost column of the
	// global-status output.  If you choose to use global_id instead of
	// source_box, Packer will skip the Vagrant initialize and add steps, and
	// simply launch the box directly using the global id.
	GlobalID string `mapstructure:"global_id" required:"true"`
	// The checksum for the .box file. The type of the checksum is specified
	// with checksum_type, documented below.
	Checksum string `mapstructure:"checksum" required:"false"`
	// The type of the checksum specified in checksum. Valid values are none,
	// md5, sha1, sha256, or sha512. Although the checksum will not be verified
	// when checksum_type is set to "none", this is not recommended since OVA
	// files can be very large and corruption does happen from time to time.
	ChecksumType string `mapstructure:"checksum_type" required:"false"`
	// if your source_box is a boxfile that we need to add to Vagrant, this is
	// the name to give it. If left blank, will default to "packer_" plus your
	// buildname.
	BoxName string `mapstructure:"box_name" required:"false"`
	// The vagrant provider.
	// This parameter is required when source_path have more than one provider,
	// or when using vagrant-cloud post-processor. Defaults to unset.
	Provider string `mapstructure:"provider" required:"false"`

	Communicator string `mapstructure:"communicator"`

	// Options for the "vagrant init" command

	// What vagrantfile to use
	VagrantfileTpl string `mapstructure:"vagrantfile_template"`
	// Whether to halt, suspend, or destroy the box when the build has
	// completed. Defaults to "halt"
	TeardownMethod string `mapstructure:"teardown_method" required:"false"`
	// What box version to use when initializing Vagrant.
	BoxVersion string `mapstructure:"box_version" required:"false"`
	// a path to a golang template for a vagrantfile. Our default template can
	// be found here. So far the only template variables available to you are
	// {{ .BoxName }} and {{ .SyncedFolder }}, which correspond to the Packer
	// options box_name and synced_folder.
	Template string `mapstructure:"template" required:"false"`

	SyncedFolder string `mapstructure:"synced_folder"`
	// Don't call "vagrant add" to add the box to your local environment; this
	// is necessary if you want to launch a box that is already added to your
	// vagrant environment.
	SkipAdd bool `mapstructure:"skip_add" required:"false"`
	// Equivalent to setting the
	// --cacert
	// option in vagrant add; defaults to unset.
	AddCACert string `mapstructure:"add_cacert" required:"false"`
	// Equivalent to setting the
	// --capath option
	// in vagrant add; defaults to unset.
	AddCAPath string `mapstructure:"add_capath" required:"false"`
	// Equivalent to setting the
	// --cert option in
	// vagrant add; defaults to unset.
	AddCert string `mapstructure:"add_cert" required:"false"`
	// Equivalent to setting the
	// --clean flag in
	// vagrant add; defaults to unset.
	AddClean bool `mapstructure:"add_clean" required:"false"`
	// Equivalent to setting the
	// --force flag in
	// vagrant add; defaults to unset.
	AddForce bool `mapstructure:"add_force" required:"false"`
	// Equivalent to setting the
	// --insecure flag in
	// vagrant add; defaults to unset.
	AddInsecure bool `mapstructure:"add_insecure" required:"false"`
	// if true, Packer will not call vagrant package to
	// package your base box into its own standalone .box file.
	SkipPackage       bool     `mapstructure:"skip_package" required:"false"`
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
			if _, err := os.Stat(b.config.SourceBox); err != nil {
				packer.MultiErrorAppend(errs,
					fmt.Errorf("Source box '%s' needs to exist at time of config validation! %v", b.config.SourceBox, err))
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
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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
			BoxName:      b.config.BoxName,
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
			Config:    &b.config.Comm,
			Host:      CommHost(),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
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
	b.runner.Run(ctx, state)

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

	return NewArtifact(b.config.Provider, b.config.OutputDir), nil
}

// Cancel.
