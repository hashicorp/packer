//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type CloneConfig,vAppConfig

package clone

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type vAppConfig struct {
	// Set values for the available vApp Properties to supply configuration parameters to a virtual machine cloned from
	// a template that came from an imported OVF or OVA file.
	//
	// -> **Note:** The only supported usage path for vApp properties is for existing user-configurable keys.
	// These generally come from an existing template that was created from an imported OVF or OVA file.
	// You cannot set values for vApp properties on virtual machines created from scratch,
	// virtual machines lacking a vApp configuration, or on property keys that do not exist.
	Properties map[string]string `mapstructure:"properties"`
}

type CloneConfig struct {
	// Name of source VM. Path is optional.
	Template string `mapstructure:"template"`
	// The size of the disk in MB.
	DiskSize int64 `mapstructure:"disk_size"`
	// Create VM as a linked clone from latest snapshot. Defaults to `false`.
	LinkedClone bool `mapstructure:"linked_clone"`
	// Set the network in which the VM will be connected to. If no network is
	// specified, `host` must be specified to allow Packer to look for the
	// available network. If the network is inside a network folder in vCenter,
	// you need to provide the full path to the network.
	Network string `mapstructure:"network"`
	// Sets a custom Mac Address to the network adapter. If set, the [network](#network) must be also specified.
	MacAddress string `mapstructure:"mac_address"`
	// VM notes.
	Notes string `mapstructure:"notes"`
	// Set the vApp Options to a virtual machine.
	// See the [vApp Options Configuration](/docs/builders/vmware/vsphere-clone#vapp-options-configuration)
	// to know the available options and how to use it.
	VAppConfig vAppConfig `mapstructure:"vapp"`
}

func (c *CloneConfig) Prepare() []error {
	var errs []error

	if c.Template == "" {
		errs = append(errs, fmt.Errorf("'template' is required"))
	}

	if c.LinkedClone == true && c.DiskSize != 0 {
		errs = append(errs, fmt.Errorf("'linked_clone' and 'disk_size' cannot be used together"))
	}

	if c.MacAddress != "" && c.Network == "" {
		errs = append(errs, fmt.Errorf("'network' is required when 'mac_address' is specified"))
	}

	return errs
}

type StepCloneVM struct {
	Config   *CloneConfig
	Location *common.LocationConfig
	Force    bool
}

func (s *StepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.VCenterDriver)
	vmPath := path.Join(s.Location.Folder, s.Location.VMName)

	err := d.PreCleanVM(ui, vmPath, s.Force)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Cloning VM...")
	template, err := d.FindVM(s.Config.Template)
	if err != nil {
		state.Put("error", fmt.Errorf("Error finding vm to clone: %s", err))
		return multistep.ActionHalt
	}

	vm, err := template.Clone(ctx, &driver.CloneConfig{
		Name:           s.Location.VMName,
		Folder:         s.Location.Folder,
		Cluster:        s.Location.Cluster,
		Host:           s.Location.Host,
		ResourcePool:   s.Location.ResourcePool,
		Datastore:      s.Location.Datastore,
		LinkedClone:    s.Config.LinkedClone,
		Network:        s.Config.Network,
		MacAddress:     s.Config.MacAddress,
		Annotation:     s.Config.Notes,
		VAppProperties: s.Config.VAppConfig.Properties,
	})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if vm == nil {
		return multistep.ActionHalt
	}
	state.Put("vm", vm)

	if s.Config.DiskSize > 0 {
		err = vm.ResizeDisk(s.Config.DiskSize)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
	common.CleanupVM(state)
}
