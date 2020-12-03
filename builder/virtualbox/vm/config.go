//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package vm

import (
	"fmt"
	"log"
	"time"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	commonsteps.HTTPConfig          `mapstructure:",squash"`
	commonsteps.FloppyConfig        `mapstructure:",squash"`
	commonsteps.CDConfig            `mapstructure:",squash"`
	bootcommand.BootConfig          `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.CommConfig           `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`
	vboxcommon.GuestAdditionsConfig `mapstructure:",squash"`
	// This is the name of the virtual machine to which the
	//  builder shall attach.
	VMName string `mapstructure:"vm_name" required:"true"`
	// Default to `null/empty`. The name of an
	//  **existing** snapshot to which the builder shall attach the VM before
	//  starting it. If no snapshot is specified the builder will simply start the
	//  VM from it's current state i.e. snapshot.
	AttachSnapshot string `mapstructure:"attach_snapshot" required:"false"`
	// Default to `null/empty`. The name of the
	//   snapshot which shall be created after all provisioners has been run by the
	//   builder. If no target snapshot is specified and `keep_registered` is set to
	//   `false` the builder will revert to the snapshot to which the VM was attached
	//   before the builder has been executed, which will revert all changes applied
	//   by the provisioners. This is handy if only an export shall be created and no
	//   further snapshot is required.
	TargetSnapshot string `mapstructure:"target_snapshot" required:"false"`
	// Defaults to `false`. If set to `true`,
	//   overwrite an existing `target_snapshot`. Otherwise the builder will yield an
	//   error if the specified target snapshot already exists.
	DeleteTargetSnapshot bool `mapstructure:"force_delete_snapshot" required:"false"`
	// Set this to `true` if you would like to keep
	//   the VM attached to the snapshot specified by `attach_snapshot`. Otherwise
	//   the builder will reset the VM to the snapshot to which the VM was attached
	//   before the builder started. Defaults to `false`.
	KeepRegistered bool `mapstructure:"keep_registered" required:"false"`
	// Defaults to `false`. When enabled, Packer will
	//   not export the VM. Useful if the builder should be applied again on the created
	//   target snapshot.
	SkipExport bool `mapstructure:"skip_export" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         vboxcommon.BuilderId, // "mitchellh.virtualbox"
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
		return nil, err
	}

	// Defaults
	if c.PostShutdownDelay == 0 {
		c.PostShutdownDelay = 2 * time.Second
	}

	// Prepare the errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.CDConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packersdk.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.CommConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(c.CommConfig.Comm.Type)...)
	errs = packersdk.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.GuestAdditionsConfig.Prepare(c.CommConfig.Comm.Type)...)
	if c.GuestAdditionsInterface == "" {
		c.GuestAdditionsInterface = "ide"
	}

	log.Printf("PostShutdownDelay: %s", c.PostShutdownDelay)

	if c.VMName == "" {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("vm_name is required"))
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
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
	} else {
		if c.AttachSnapshot != "" && c.TargetSnapshot != "" && c.AttachSnapshot == c.TargetSnapshot {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Attach snapshot %s and target snapshot %s cannot be the same", c.AttachSnapshot, c.TargetSnapshot))
		}
		snapshotTree, err := driver.LoadSnapshots(c.VMName)
		log.Printf("")
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Failed to load snapshots for VM %s: %s", c.VMName, err))
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
					errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("No snapshots defined on VM %s. Unable to attach to %s", c.VMName, c.AttachSnapshot))
				} else {
					snapshots := snapshotTree.GetSnapshotsByName(c.AttachSnapshot)
					if 0 >= len(snapshots) {
						errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Snapshot %s does not exist on VM %s", c.AttachSnapshot, c.VMName))
					} else if 1 < len(snapshots) {
						errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Multiple Snapshots with name %s exist on VM %s", c.AttachSnapshot, c.VMName))
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
						errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s cannot be the same as the snapshot to which the builder shall attach: %s", c.TargetSnapshot, attachSnapshot.Name))
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
								errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s already exists and is not a direct child of %s", c.TargetSnapshot, attachSnapshot.Name))
							} else if !c.DeleteTargetSnapshot {
								errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Target snapshot %s already exists as direct child of %s for VM %s. Use force_delete_snapshot = true to overwrite snapshot",
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
		return warnings, errs
	}

	return warnings, nil
}
