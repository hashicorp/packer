//go:generate struct-markdown

package iso

import (
	"context"
	"errors"
	"fmt"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "rickard-von-essen.parallels"

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig                 `mapstructure:",squash"`
	common.HTTPConfig                   `mapstructure:",squash"`
	common.ISOConfig                    `mapstructure:",squash"`
	common.FloppyConfig                 `mapstructure:",squash"`
	bootcommand.BootConfig              `mapstructure:",squash"`
	parallelscommon.OutputConfig        `mapstructure:",squash"`
	parallelscommon.HWConfig            `mapstructure:",squash"`
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlPostConfig    `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`
	parallelscommon.ShutdownConfig      `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`
	parallelscommon.ToolsConfig         `mapstructure:",squash"`
	// The size, in megabytes, of the hard disk to create
    // for the VM. By default, this is 40000 (about 40 GB).
	DiskSize           uint     `mapstructure:"disk_size" required:"false"`
	// The type for image file based virtual disk drives,
    // defaults to expand. Valid options are expand (expanding disk) that the
    // image file is small initially and grows in size as you add data to it, and
    // plain (plain disk) that the image file has a fixed size from the moment it
    // is created (i.e the space is allocated for the full drive). Plain disks
    // perform faster than expanding disks. skip_compaction will be set to true
    // automatically for plain disks.
	DiskType           string   `mapstructure:"disk_type" required:"false"`
	// The guest OS type being installed. By default
    // this is "other", but you can get dramatic performance improvements by
    // setting this to the proper value. To view all available values for this run
    // prlctl create x --distribution list. Setting the correct value hints to
    // Parallels Desktop how to optimize the virtual hardware to work best with
    // that operating system.
	GuestOSType        string   `mapstructure:"guest_os_type" required:"false"`
	// The type of controller that the hard
    // drives are attached to, defaults to "sata". Valid options are "sata", "ide",
    // and "scsi".
	HardDriveInterface string   `mapstructure:"hard_drive_interface" required:"false"`
	// A list of which interfaces on the
    // host should be searched for a IP address. The first IP address found on one
    // of these will be used as {{ .HTTPIP }} in the boot_command. Defaults to
    // ["en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7", "en8", "en9",
    // "ppp0", "ppp1", "ppp2"].
	HostInterfaces     []string `mapstructure:"host_interfaces" required:"false"`
	// Virtual disk image is compacted at the end of
    // the build process using prl_disk_tool utility (except for the case that
    // disk_type is set to plain). In certain rare cases, this might corrupt
    // the resulting disk image. If you find this to be the case, you can disable
    // compaction using this configuration value.
	SkipCompaction     bool     `mapstructure:"skip_compaction" required:"false"`
	// This is the name of the PVM directory for the new
    // virtual machine, without the file extension. By default this is
    // "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName             string   `mapstructure:"vm_name" required:"false"`

	ctx interpolate.Context
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"prlctl",
				"prlctl_post",
				"parallels_tools_guest_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(
		errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.HWConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.PrlctlConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.PrlctlPostConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.PrlctlVersionConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ToolsConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.BootConfig.Prepare(&b.config.ctx)...)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.DiskType == "" {
		b.config.DiskType = "expand"
	}

	if b.config.HardDriveInterface == "" {
		b.config.HardDriveInterface = "sata"
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "other"
	}

	if len(b.config.HostInterfaces) == 0 {
		b.config.HostInterfaces = []string{"en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7",
			"en8", "en9", "ppp0", "ppp1", "ppp2"}
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.DiskType != "expand" && b.config.DiskType != "plain" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("disk_type can only be expand, or plain"))
	}

	if b.config.DiskType == "plain" && !b.config.SkipCompaction {
		b.config.SkipCompaction = true
		warnings = append(warnings,
			"'skip_compaction' is enforced to be true for plain disks.")
	}

	if b.config.HardDriveInterface != "ide" && b.config.HardDriveInterface != "sata" && b.config.HardDriveInterface != "scsi" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("hard_drive_interface can only be ide, sata, or scsi"))
	}

	// Warnings
	if b.config.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Parallels
	driver, err := parallelscommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Parallels driver: %s", err)
	}

	steps := []multistep.Step{
		&parallelscommon.StepPrepareParallelsTools{
			ParallelsToolsFlavor: b.config.ParallelsToolsFlavor,
			ParallelsToolsMode:   b.config.ParallelsToolsMode,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			Extension:    b.config.TargetExtension,
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
			Url:          b.config.ISOUrls,
		},
		&parallelscommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
		},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		new(stepCreateVM),
		new(stepCreateDisk),
		new(stepSetBootOrder),
		new(stepAttachISO),
		&parallelscommon.StepAttachParallelsTools{
			ParallelsToolsMode: b.config.ParallelsToolsMode,
		},
		new(parallelscommon.StepAttachFloppy),
		&parallelscommon.StepPrlctl{
			Commands: b.config.Prlctl,
			Ctx:      b.config.ctx,
		},
		&parallelscommon.StepRun{},
		&parallelscommon.StepTypeBootCommand{
			BootWait:       b.config.BootWait,
			BootCommand:    b.config.FlatBootCommand(),
			HostInterfaces: b.config.HostInterfaces,
			VMName:         b.config.VMName,
			Ctx:            b.config.ctx,
			GroupInterval:  b.config.BootConfig.BootGroupInterval,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      parallelscommon.CommHost,
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
		&parallelscommon.StepUploadVersion{
			Path: b.config.PrlctlVersionFile,
		},
		&parallelscommon.StepUploadParallelsTools{
			ParallelsToolsFlavor:    b.config.ParallelsToolsFlavor,
			ParallelsToolsGuestPath: b.config.ParallelsToolsGuestPath,
			ParallelsToolsMode:      b.config.ParallelsToolsMode,
			Ctx:                     b.config.ctx,
		},
		new(common.StepProvision),
		&common.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},
		&parallelscommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&parallelscommon.StepPrlctl{
			Commands: b.config.PrlctlPost,
			Ctx:      b.config.ctx,
		},
		&parallelscommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
		},
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
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

	return parallelscommon.NewArtifact(b.config.OutputDir)
}
