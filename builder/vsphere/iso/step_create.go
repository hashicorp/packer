//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type NIC,CreateConfig,DiskConfig

package iso

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type NIC struct {
	// Set network VM will be connected to.
	Network string `mapstructure:"network"`
	// Set VM network card type. Example `vmxnet3`.
	NetworkCard string `mapstructure:"network_card" required:"true"`
	// Set network card MAC address
	MacAddress string `mapstructure:"mac_address"`
	// Enable DirectPath I/O passthrough
	Passthrough *bool `mapstructure:"passthrough"`
}

type DiskConfig struct {
	// Set the size of the disk
	DiskSize int64 `mapstructure:"disk_size" required:"true"`
	// Enable VMDK thin provisioning for VM. Defaults to `false`.
	DiskThinProvisioned bool `mapstructure:"disk_thin_provisioned"`
	// Enable VMDK eager scrubbing for VM. Defaults to `false`.
	DiskEagerlyScrub bool `mapstructure:"disk_eagerly_scrub"`
}

type CreateConfig struct {
	// Set VM hardware version. Defaults to the most current VM hardware
	// version supported by vCenter. See
	// [VMWare article 1003746](https://kb.vmware.com/s/article/1003746) for
	// the full list of supported VM hardware versions.
	Version uint `mapstructure:"vm_version"`
	// Set VM OS type. Defaults to `otherGuest`. See [
	// here](https://code.vmware.com/apis/358/vsphere/doc/vim.vm.GuestOsDescriptor.GuestOsIdentifier.html)
	// for a full list of possible values.
	GuestOSType string `mapstructure:"guest_os_type"`
	// Set the Firmware at machine creation. Example `efi`. Defaults to `bios`.
	Firmware string `mapstructure:"firmware"`
	// Set VM disk controller type. Example `pvscsi`.
	DiskControllerType string `mapstructure:"disk_controller_type"`
	// The size of the disk in MB.
	DiskSize int64 `mapstructure:"disk_size"`
	// Enable VMDK thin provisioning for VM. Defaults to `false`.
	DiskThinProvisioned bool `mapstructure:"disk_thin_provisioned"`
	// Enable VMDK eager scrubbing for VM. Defaults to `false`.
	DiskEagerlyScrub bool `mapstructure:"disk_eagerly_scrub"`
	// A collection of one or more disks to be provisioned along with the VM.
	Storage []DiskConfig `mapstructure:"storage"`
	// Set network VM will be connected to.
	Network string `mapstructure:"network"`
	// Set VM network card type. Example `vmxnet3`.
	NetworkCard string `mapstructure:"network_card"`
	// Network adapters
	NICs []NIC `mapstructure:"network_adapters"`
	// Create USB controller for virtual machine. Defaults to `false`.
	USBController bool `mapstructure:"usb_controller"`
	// VM notes.
	Notes string `mapstructure:"notes"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	if c.DiskSize == 0 {
		errs = append(errs, fmt.Errorf("'disk_size' is required"))
	}

	if c.GuestOSType == "" {
		c.GuestOSType = "otherGuest"
	}

	if c.Firmware != "" && c.Firmware != "bios" && c.Firmware != "efi" {
		errs = append(errs, fmt.Errorf("'firmware' must be 'bios' or 'efi'"))
	}

	return errs
}

type StepCreateVM struct {
	Config   *CreateConfig
	Location *common.LocationConfig
	Force    bool
}

func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	vm, err := d.FindVM(s.Location.VMName)

	if s.Force == false && err == nil {
		state.Put("error", fmt.Errorf("%s already exists, you can use -force flag to destroy it: %v", s.Location.VMName, err))
		return multistep.ActionHalt
	} else if s.Force == true && err == nil {
		ui.Say(fmt.Sprintf("the vm/template %s already exists, but deleting it due to -force flag", s.Location.VMName))
		err := vm.Destroy()
		if err != nil {
			state.Put("error", fmt.Errorf("error destroying %s: %v", s.Location.VMName, err))
		}
	}

	ui.Say("Creating VM...")

	// add network/network card an the first nic for backwards compatibility in the type is defined
	var networkCards []driver.NIC
	if s.Config.NetworkCard != "" {
		networkCards = append(networkCards, driver.NIC{
			NetworkCard: s.Config.NetworkCard,
			Network:     s.Config.Network})
	}
	for _, nic := range s.Config.NICs {
		networkCards = append(networkCards, driver.NIC{
			Network:     nic.Network,
			NetworkCard: nic.NetworkCard,
			MacAddress:  nic.MacAddress,
			Passthrough: nic.Passthrough,
		})
	}

	// add disk as the first drive for backwards compatibility if the type is defined
	var disks []driver.Disk
	if s.Config.DiskSize != 0 {
		disks = append(disks, driver.Disk{
			DiskSize:            s.Config.DiskSize,
			DiskEagerlyScrub:    s.Config.DiskEagerlyScrub,
			DiskThinProvisioned: s.Config.DiskThinProvisioned,
		})
	}
	for _, disk := range s.Config.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
		})
	}

	vm, err = d.CreateVM(&driver.CreateConfig{
		DiskControllerType: s.Config.DiskControllerType,
		Storage:            disks,
		Annotation:         s.Config.Notes,
		Name:               s.Location.VMName,
		Folder:             s.Location.Folder,
		Cluster:            s.Location.Cluster,
		Host:               s.Location.Host,
		ResourcePool:       s.Location.ResourcePool,
		Datastore:          s.Location.Datastore,
		GuestOS:            s.Config.GuestOSType,
		NICs:               networkCards,
		USBController:      s.Config.USBController,
		Version:            s.Config.Version,
		Firmware:           s.Config.Firmware,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating vm: %v", err))
		return multistep.ActionHalt
	}
	state.Put("vm", vm)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(*driver.VirtualMachine)

	ui.Say("Destroying VM...")
	err := vm.Destroy()
	if err != nil {
		ui.Error(err.Error())
	}
}
