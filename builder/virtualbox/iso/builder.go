package iso

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.ExportOpts           `mapstructure:",squash"`
	vboxcommon.FloppyConfig         `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.SSHConfig            `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxManagePostConfig `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`

	BootCommand          []string `mapstructure:"boot_command"`
	DiskSize             uint     `mapstructure:"disk_size"`
	GuestAdditionsMode   string   `mapstructure:"guest_additions_mode"`
	GuestAdditionsPath   string   `mapstructure:"guest_additions_path"`
	GuestAdditionsURL    string   `mapstructure:"guest_additions_url"`
	GuestAdditionsSHA256 string   `mapstructure:"guest_additions_sha256"`
	GuestOSType          string   `mapstructure:"guest_os_type"`
	HardDriveInterface   string   `mapstructure:"hard_drive_interface"`
	ISOChecksum          string   `mapstructure:"iso_checksum"`
	ISOChecksumType      string   `mapstructure:"iso_checksum_type"`
	ISOInterface         string   `mapstructure:"iso_interface"`
	ISOUrls              []string `mapstructure:"iso_urls"`
	VMName               string   `mapstructure:"vm_name"`

	RawSingleISOUrl string `mapstructure:"iso_url"`

	ctx interpolate.Context
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"guest_additions_path",
				"guest_additions_url",
				"vboxmanage",
				"vboxmanage_post",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.ExportConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ExportOpts.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(
		errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxManageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxManagePostConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxVersionConfig.Prepare(&b.config.ctx)...)
	warnings := make([]string, 0)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.GuestAdditionsMode == "" {
		b.config.GuestAdditionsMode = "upload"
	}

	if b.config.GuestAdditionsPath == "" {
		b.config.GuestAdditionsPath = "VBoxGuestAdditions.iso"
	}

	if b.config.HardDriveInterface == "" {
		b.config.HardDriveInterface = "ide"
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "Other"
	}

	if b.config.ISOInterface == "" {
		b.config.ISOInterface = "ide"
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", b.config.PackerBuildName)
	}

	if b.config.HardDriveInterface != "ide" && b.config.HardDriveInterface != "sata" && b.config.HardDriveInterface != "scsi" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("hard_drive_interface can only be ide, sata, or scsi"))
	}

	if b.config.ISOChecksumType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		b.config.ISOChecksumType = strings.ToLower(b.config.ISOChecksumType)
		if b.config.ISOChecksumType != "none" {
			if b.config.ISOChecksum == "" {
				errs = packer.MultiErrorAppend(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
			} else {
				b.config.ISOChecksum = strings.ToLower(b.config.ISOChecksum)
			}

			if h := common.HashForType(b.config.ISOChecksumType); h == nil {
				errs = packer.MultiErrorAppend(
					errs,
					fmt.Errorf("Unsupported checksum type: %s", b.config.ISOChecksumType))
			}
		}
	}

	if b.config.ISOInterface != "ide" && b.config.ISOInterface != "sata" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("iso_interface can only be ide or sata"))
	}

	if b.config.RawSingleISOUrl == "" && len(b.config.ISOUrls) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
	} else if b.config.RawSingleISOUrl != "" && len(b.config.ISOUrls) > 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
	} else if b.config.RawSingleISOUrl != "" {
		b.config.ISOUrls = []string{b.config.RawSingleISOUrl}
	}

	for i, url := range b.config.ISOUrls {
		b.config.ISOUrls[i], err = common.DownloadableURL(url)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed to parse iso_url %d: %s", i+1, err))
		}
	}

	validMode := false
	validModes := []string{
		vboxcommon.GuestAdditionsModeDisable,
		vboxcommon.GuestAdditionsModeAttach,
		vboxcommon.GuestAdditionsModeUpload,
	}

	for _, mode := range validModes {
		if b.config.GuestAdditionsMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("guest_additions_mode is invalid. Must be one of: %v", validModes))
	}

	if b.config.GuestAdditionsSHA256 != "" {
		b.config.GuestAdditionsSHA256 = strings.ToLower(b.config.GuestAdditionsSHA256)
	}

	// Warnings
	if b.config.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

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
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// Create the driver that we'll use to communicate with VirtualBox
	driver, err := vboxcommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	steps := []multistep.Step{
		&vboxcommon.StepDownloadGuestAdditions{
			GuestAdditionsMode:   b.config.GuestAdditionsMode,
			GuestAdditionsURL:    b.config.GuestAdditionsURL,
			GuestAdditionsSHA256: b.config.GuestAdditionsSHA256,
			Ctx:                  b.config.ctx,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
			Extension:    "iso",
		},
		&vboxcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&vboxcommon.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		new(vboxcommon.StepSuppressMessages),
		new(stepCreateVM),
		new(stepCreateDisk),
		new(stepAttachISO),
		&vboxcommon.StepAttachGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
		},
		new(vboxcommon.StepAttachFloppy),
		&vboxcommon.StepForwardSSH{
			GuestPort:      uint(b.config.SSHConfig.Comm.SSHPort),
			HostPortMin:    b.config.SSHHostPortMin,
			HostPortMax:    b.config.SSHHostPortMax,
			SkipNatMapping: b.config.SSHSkipNatMapping,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManage,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepRun{
			BootWait: b.config.BootWait,
			Headless: b.config.Headless,
		},
		&vboxcommon.StepTypeBootCommand{
			BootCommand: b.config.BootCommand,
			VMName:      b.config.VMName,
			Ctx:         b.config.ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      vboxcommon.CommHost,
			SSHConfig: vboxcommon.SSHConfigFunc(b.config.SSHConfig),
			SSHPort:   vboxcommon.SSHPort,
		},
		&vboxcommon.StepUploadVersion{
			Path: b.config.VBoxVersionFile,
		},
		&vboxcommon.StepUploadGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
			GuestAdditionsPath: b.config.GuestAdditionsPath,
			Ctx:                b.config.ctx,
		},
		new(common.StepProvision),
		&vboxcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		new(vboxcommon.StepRemoveDevices),
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManagePost,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepExport{
			Format:         b.config.Format,
			OutputDir:      b.config.OutputDir,
			ExportOpts:     b.config.ExportOpts.ExportOpts,
			SkipNatMapping: b.config.SSHSkipNatMapping,
		},
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

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

	return vboxcommon.NewArtifact(b.config.OutputDir)
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
