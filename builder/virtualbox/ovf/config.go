package ovf

import (
	"fmt"
	"os"
	"strings"

	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.ExportOpts           `mapstructure:",squash"`
	vboxcommon.FloppyConfig         `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.SSHConfig            `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxManagePostConfig `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`

	BootCommand          []string `mapstructure:"boot_command"`
	SourcePath           string   `mapstructure:"source_path"`
	GuestAdditionsMode   string   `mapstructure:"guest_additions_mode"`
	GuestAdditionsPath   string   `mapstructure:"guest_additions_path"`
	GuestAdditionsURL    string   `mapstructure:"guest_additions_url"`
	GuestAdditionsSHA256 string   `mapstructure:"guest_additions_sha256"`
	VMName               string   `mapstructure:"vm_name"`
	ImportOpts           string   `mapstructure:"import_opts"`
	ImportFlags          []string `mapstructure:"import_flags"`

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

	if c.VMName == "" {
		c.VMName = fmt.Sprintf(
			"packer-%s-%d", c.PackerBuildName, interpolate.InitTime.Unix())
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportOpts.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManagePostConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(&c.ctx)...)

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		if _, err := os.Stat(c.SourcePath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("source_path is invalid: %s", err))
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
