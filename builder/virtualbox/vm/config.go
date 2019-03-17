package vm

import (
	"fmt"
	"log"
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

	GuestAdditionsMode   string `mapstructure:"guest_additions_mode"`
	GuestAdditionsPath   string `mapstructure:"guest_additions_path"`
	GuestAdditionsSHA256 string `mapstructure:"guest_additions_sha256"`
	GuestAdditionsURL    string `mapstructure:"guest_additions_url"`
	VMName               string `mapstructure:"vm_name"`
	AttachSnapshot       string `mapstructure:"attach_snapshot"`
	TargetSnapshot       string `mapstructure:"target_snapshot"`
	KeepRegistered       bool   `mapstructure:"keep_registered"`
	SkipExport           bool   `mapstructure:"skip_export"`

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

	if c.VMName == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("vm_name is required"))
	}

	if c.TargetSnapshot == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("target_snapshot is required"))
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
	driver, err := vboxcommon.NewDriver()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
	} else {
		snapshotTree, err := driver.LoadSnapshots(c.VMName)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
		} else {
			if c.AttachSnapshot != "" && c.TargetSnapshot != "" && c.AttachSnapshot == c.TargetSnapshot {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("Attach snapshot %s and target snapshot %s cannot be the same", c.AttachSnapshot, c.TargetSnapshot))
			}
			attachSnapshot := snapshotTree.GetCurrentSnapshot()
			if c.AttachSnapshot != "" {
				snapshots := snapshotTree.GetSnapshotsByName(c.AttachSnapshot)
				if 0 >= len(snapshots) {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("Snapshot %s does not exist on with VM %s", c.AttachSnapshot, c.VMName))
				} else if 1 < len(snapshots) {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("Multiple Snapshots %s exist on with VM %s", c.AttachSnapshot, c.VMName))
				} else {
					attachSnapshot = snapshots[0]
				}
			}
			if c.TargetSnapshot != "" {
				snapshots := snapshotTree.GetSnapshotsByName(c.TargetSnapshot)
				if 0 >= len(snapshots) {
					isChild := false
					for _, snapshot := range snapshots {
						log.Printf("Checking if target snaphot %v is child of %s")
						isChild = nil != snapshot.Parent && snapshot.Parent.UUID == attachSnapshot.UUID
					}
					if !isChild {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s already exists and is not a direct child of %s", c.TargetSnapshot, attachSnapshot.Name))
					}
				}
			}
		}
	}
	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
