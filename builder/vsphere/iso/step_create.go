//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type NIC,CreateConfig

package iso

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

// Defines a Network Adapter
//
// Example that creates two network adapters:
//
// In JSON:
// ```json
//   "network_adapters": [
//     {
//       "network": "VM Network",
//       "network_card": "vmxnet3"
//     },
//     {
//       "network": "OtherNetwork",
//       "network_card": "vmxnet3"
//     }
//   ],
// ```
// In HCL2:
// ```hcl
//   network_adapters {
//       network = "VM Network"
//       network_card = "vmxnet3"
//   }
//   network_adapters {
//       network = "OtherNetwork"
//       network_card = "vmxnet3"
//   }
// ```
type NIC struct {
	// Set the network in which the VM will be connected to. If no network is
	// specified, `host` must be specified to allow Packer to look for the
	// available network. If the network is inside a network folder in vCenter,
	// you need to provide the full path to the network.
	Network string `mapstructure:"network"`
	// Set VM network card type. Example `vmxnet3`.
	NetworkCard string `mapstructure:"network_card" required:"true"`
	// Set network card MAC address
	MacAddress string `mapstructure:"mac_address"`
	// Enable DirectPath I/O passthrough
	Passthrough *bool `mapstructure:"passthrough"`
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
	GuestOSType   string               `mapstructure:"guest_os_type"`
	StorageConfig common.StorageConfig `mapstructure:",squash"`
	// Network adapters
	NICs []NIC `mapstructure:"network_adapters"`
	// Create USB controllers for the virtual machine. "usb" for a usb 2.0 controller. "xhci" for a usb 3.0 controller. There can only be at most one of each.
	USBController []string `mapstructure:"usb_controller"`
	// VM notes.
	Notes string `mapstructure:"notes"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	if len(c.StorageConfig.DiskControllerType) == 0 {
		c.StorageConfig.DiskControllerType = append(c.StorageConfig.DiskControllerType, "")
	}

	// there should be at least one
	if len(c.StorageConfig.Storage) == 0 {
		errs = append(errs, fmt.Errorf("no storage devices have been defined"))
	}
	errs = append(errs, c.StorageConfig.Prepare()...)

	if c.GuestOSType == "" {
		c.GuestOSType = "otherGuest"
	}

	usbCount := 0
	xhciCount := 0

	for i, s := range c.USBController {
		switch s {
		// 1 and true for backwards compatibility
		case "usb", "1", "true":
			usbCount++
		case "xhci":
			xhciCount++
		// 0 and false for backwards compatibility
		case "false", "0":
			continue
		default:
			errs = append(errs, fmt.Errorf("usb_controller[%d] references an unknown usb controller", i))
		}
	}
	if usbCount > 1 || xhciCount > 1 {
		errs = append(errs, fmt.Errorf("there can only be one usb controller and one xhci controller"))
	}

	return errs
}

type StepCreateVM struct {
	Config   *CreateConfig
	Location *common.LocationConfig
	Force    bool
}

func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(driver.Driver)
	vmPath := path.Join(s.Location.Folder, s.Location.VMName)

	err := d.PreCleanVM(ui, vmPath, s.Force)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Creating VM...")

	// add network/network card an the first nic for backwards compatibility in the type is defined
	var networkCards []driver.NIC
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
	for _, disk := range s.Config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	vm, err := d.CreateVM(&driver.CreateConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: s.Config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:    s.Config.Notes,
		Name:          s.Location.VMName,
		Folder:        s.Location.Folder,
		Cluster:       s.Location.Cluster,
		Host:          s.Location.Host,
		ResourcePool:  s.Location.ResourcePool,
		Datastore:     s.Location.Datastore,
		GuestOS:       s.Config.GuestOSType,
		NICs:          networkCards,
		USBController: s.Config.USBController,
		Version:       s.Config.Version,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating vm: %v", err))
		return multistep.ActionHalt
	}
	state.Put("vm", vm)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	common.CleanupVM(state)
}
