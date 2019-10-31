//go:generate mapstructure-to-hcl2 -type Config

package vm

import (
	"fmt"
	"log"
	"strings"
	"time"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig          `mapstructure:",squash"`
	common.HTTPConfig            `mapstructure:",squash"`
	common.FloppyConfig          `mapstructure:",squash"`
	bootcommand.BootConfig       `mapstructure:",squash"`
	vboxcommon.ExportConfig      `mapstructure:",squash"`
	vboxcommon.OutputConfig      `mapstructure:",squash"`
	vboxcommon.RunConfig         `mapstructure:",squash"`
	vboxcommon.SSHConfig         `mapstructure:",squash"`
	vboxcommon.ShutdownConfig    `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig  `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig `mapstructure:",squash"`

	GuestAdditionsMode   string `mapstructure:"guest_additions_mode"`
	GuestAdditionsPath   string `mapstructure:"guest_additions_path"`
	GuestAdditionsSHA256 string `mapstructure:"guest_additions_sha256"`
	GuestAdditionsURL    string `mapstructure:"guest_additions_url"`
	VMName               string `mapstructure:"vm_name"`
	AttachSnapshot       string `mapstructure:"attach_snapshot"`
	TargetSnapshot       string `mapstructure:"target_snapshot"`
	DeleteTargetSnapshot bool   `mapstructure:"force_delete_snapshot"`
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

	if c.PostShutdownDelay == 0 {
		c.PostShutdownDelay = 2 * time.Second
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)

	log.Printf("PostShutdownDelay: %s", c.PostShutdownDelay)

	if c.VMName == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("vm_name is required"))
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
	if c.TargetSnapshot == "" && c.SkipExport {
		warnings = append(warnings,
			"No target snapshot is specified (target_snapshot empty) and no export will be created (skip_export=true).\n"+
				"You might lose all changes applied by this run, the next time you execute packer.")
	}

	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	driver, err := vboxcommon.NewDriver()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
	} else {
		if c.AttachSnapshot != "" && c.TargetSnapshot != "" && c.AttachSnapshot == c.TargetSnapshot {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Attach snapshot %s and target snapshot %s cannot be the same", c.AttachSnapshot, c.TargetSnapshot))
		}
		snapshotTree, err := driver.LoadSnapshots(c.VMName)
		log.Printf("")
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed to load snapshots for VM %s: %s", c.VMName, err))
		} else {
			log.Printf("Snapshots loaded from VM %s", c.VMName)

			var attachSnapshot *vboxcommon.VBoxSnapshot
			if nil != snapshotTree {
				attachSnapshot = snapshotTree.GetCurrentSnapshot()
				log.Printf("VM %s is currently attached to snapshot: %s/%s", c.VMName, attachSnapshot.Name, attachSnapshot.UUID)
			}
			if c.AttachSnapshot != "" {
				log.Printf("Checking configuration attach_snapshot [%s]", c.AttachSnapshot)
				if nil == snapshotTree {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("No snapshots defined on VM %s. Unable to attach to %s", c.VMName, c.AttachSnapshot))
				} else {
					snapshots := snapshotTree.GetSnapshotsByName(c.AttachSnapshot)
					if 0 >= len(snapshots) {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("Snapshot %s does not exist on VM %s", c.AttachSnapshot, c.VMName))
					} else if 1 < len(snapshots) {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("Multiple Snapshots with name %s exist on VM %s", c.AttachSnapshot, c.VMName))
					} else {
						attachSnapshot = snapshots[0]
					}
				}
			}
			if c.TargetSnapshot != "" {
				log.Printf("Checking configuration target_snapshot [%s]", c.TargetSnapshot)
				if nil == snapshotTree {
					log.Printf("Currently no snapshots defined in VM %s", c.VMName)
				} else {
					if c.TargetSnapshot == attachSnapshot.Name {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s cannot be the same as the snapshot to which the builder shall attach: %s", c.TargetSnapshot, attachSnapshot.Name))
					} else {
						snapshots := snapshotTree.GetSnapshotsByName(c.TargetSnapshot)
						if 0 < len(snapshots) {
							if nil == attachSnapshot {
								panic("Internal error. Expecting a handle to a VBoxSnapshot")
							}
							isChild := false
							for _, snapshot := range snapshots {
								log.Printf("Checking if target snaphot %s/%s is child of %s/%s", snapshot.Name, snapshot.UUID, attachSnapshot.Name, attachSnapshot.UUID)
								isChild = nil != snapshot.Parent && snapshot.Parent.UUID == attachSnapshot.UUID
							}
							if !isChild {
								errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s already exists and is not a direct child of %s", c.TargetSnapshot, attachSnapshot.Name))
							} else if !c.DeleteTargetSnapshot {
								errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s already exists as direct child of %s for VM %s. Use force_delete_snapshot = true to overwrite snapshot",
									c.TargetSnapshot,
									attachSnapshot.Name,
									c.VMName))
							}
						} else {
							log.Printf("No snapshot with name %s defined in VM %s", c.TargetSnapshot, c.VMName)
						}
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
