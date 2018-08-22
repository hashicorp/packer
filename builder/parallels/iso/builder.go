package iso

import (
	"errors"
	"fmt"
	"log"

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
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlPostConfig    `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`
	parallelscommon.ShutdownConfig      `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`
	parallelscommon.ToolsConfig         `mapstructure:",squash"`

	DiskSize           uint     `mapstructure:"disk_size"`
	DiskType           string   `mapstructure:"disk_type"`
	GuestOSType        string   `mapstructure:"guest_os_type"`
	HardDriveInterface string   `mapstructure:"hard_drive_interface"`
	HostInterfaces     []string `mapstructure:"host_interfaces"`
	SkipCompaction     bool     `mapstructure:"skip_compaction"`
	VMName             string   `mapstructure:"vm_name"`

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

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
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
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

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

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
