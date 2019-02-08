Gpackage ovf

import (
	"fmt"
	"strings"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	common.HTTPConfig               `mapstructure:",squash"`
	common.FloppyConfig             `mapstructure:",squash"`
	bootcommand.BootConfig          `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.ExportOpts           `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.SSHConfig            `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxManagePostConfig `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`

	Checksum                string   `mapstructure:"checksum"`
	ChecksumType            string   `mapstructure:"checksum_type"`
	GuestAdditionsMode      string   `mapstructure:"guest_additions_mode"`
	GuestAdditionsPath      string   `mapstructure:"guest_additions_path"`
	GuestAdditionsInterface string   `mapstructure:"guest_additions_interface"`
	GuestAdditionsSHA256    string   `mapstructure:"guest_additions_sha256"`
	GuestAdditionsURL       string   `mapstructure:"guest_additions_url"`
	ImportFlags             []string `mapstructure:"import_flags"`
	ImportOpts              string   `mapstructure:"import_opts"`
	SourcePath              string   `mapstructure:"source_path"`
	TargetPath              string   `mapstructure:"target_path"`
	VMName                  string   `mapstructure:"vm_name"`
	KeepRegistered          bool     `mapstructure:"keep_registered"`
	SkipExport              bool     `mapstructure:"skip_export"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
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
		return nil, nil, err
	}

	// Defaults
	if c.GuestAdditionsMode == "" {
		c.GuestAdditionsMode = "upload"
	}

	if c.GuestAdditionsPath == "" {
		c.GuestAdditionsPath = "VBoxGuestAdditions.iso"
	}
	if b.config.GuestAdditionsInterface == "" {
		b.config.GuestAdditionsInterface = "ide"
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf(
			"packer-%s-%d", c.PackerBuildName, interpolate.InitTime.Unix())
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportOpts.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManagePostConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)

	c.ChecksumType = strings.ToLower(c.ChecksumType)
	c.Checksum = strings.ToLower(c.Checksum)

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		c.SourcePath, err = common.ValidatedURL(c.SourcePath)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is invalid: %s", err))
		}
		fileOK := common.FileExistsLocally(c.SourcePath)
		if !fileOK {
			packer.MultiErrorAppend(errs,
				fmt.Errorf("Source file '%s' needs to exist at time of config validation!", c.SourcePath))
		}

	}

	validMode := false
	validModes := []string{
		vboxcommon.GuestAdditionsModeDisable,
		vboxcommon.GuestAdditionsModeAttach,
		vboxcommon.GuestAdditionsModeUpload,
	}

	for _, mode := range validModes {
		if c.GuestAdditionsMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("guest_additions_mode is invalid. Must be one of: %v", validModes))
	}

	if c.GuestAdditionsSHA256 != "" {
		c.GuestAdditionsSHA256 = strings.ToLower(c.GuestAdditionsSHA256)
	}

	// Warnings
	var warnings []string
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	// TODO: Write a packer fix and just remove import_opts
	if c.ImportOpts != "" {
		c.ImportFlags = append(c.ImportFlags, "--options", c.ImportOpts)
	}

	return c, warnings, nil
}
