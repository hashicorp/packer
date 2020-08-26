//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type NIC,CreateConfig,DiskConfig

package iso

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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

// Defines the disk storage for a VM.
//
// Example that will create a 15GB and a 20GB disk on the VM. The second disk will be thin provisioned:
//
// In JSON:
// ```json
//   "storage": [
//     {
//       "disk_size": 15000
//     },
//     {
//       "disk_size": 20000,
//       "disk_thin_provisioned": true
//     }
//   ],
// ```
// In HCL2:
// ```hcl
//   storage {
//       disk_size = 15000
//   }
//   storage {
//       disk_size = 20000
//       disk_thin_provisioned = true
//   }
// ```
//
// Example that creates 2 pvscsi controllers and adds 2 disks to each one:
//
// In JSON:
// ```json
//   "disk_controller_type": ["pvscsi", "pvscsi"],
//   "storage": [
//     {
//       "disk_size": 15000,
//       "disk_controller_index": 0
//     },
//     {
//       "disk_size": 15000,
//       "disk_controller_index": 0
//     },
//     {
//       "disk_size": 15000,
//       "disk_controller_index": 1
//     },
//     {
//       "disk_size": 15000,
//       "disk_controller_index": 1
//     }
//   ],
// ```
//
// In HCL2:
// ```hcl
//   disk_controller_type = ["pvscsi", "pvscsi"]
//   storage {
//      disk_size = 15000,
//      disk_controller_index = 0
//   }
//   storage {
//      disk_size = 15000
//      disk_controller_index = 0
//   }
//   storage {
//      disk_size = 15000
//      disk_controller_index = 1
//   }
//   storage {
//      disk_size = 15000
//      disk_controller_index = 1
//   }
// ```
type DiskConfig struct {
	// The size of the disk in MB.
	DiskSize int64 `mapstructure:"disk_size" required:"true"`
	// Enable VMDK thin provisioning for VM. Defaults to `false`.
	DiskThinProvisioned bool `mapstructure:"disk_thin_provisioned"`
	// Enable VMDK eager scrubbing for VM. Defaults to `false`.
	DiskEagerlyScrub bool `mapstructure:"disk_eagerly_scrub"`
	// The assigned disk controller. Defaults to the first one (0)
	DiskControllerIndex int `mapstructure:"disk_controller_index"`
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
	// Set VM disk controller type. Example `lsilogic`, pvscsi`, or `scsi`. Use a list to define additional controllers. Defaults to `lsilogic`
	DiskControllerType []string `mapstructure:"disk_controller_type"`
	// A collection of one or more disks to be provisioned along with the VM.
	Storage []DiskConfig `mapstructure:"storage"`
	// Network adapters
	NICs []NIC `mapstructure:"network_adapters"`
	// Create USB controllers for the virtual machine. "usb" for a usb 2.0 controller. "xhci" for a usb 3.0 controller. There can only be at most one of each.
	USBController []string `mapstructure:"usb_controller"`
	// VM notes.
	Notes string `mapstructure:"notes"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	// there should be at least one
	if len(c.DiskControllerType) == 0 {
		c.DiskControllerType = append(c.DiskControllerType, "")
	}

	if len(c.Storage) > 0 {
		for i, storage := range c.Storage {
			if storage.DiskSize == 0 {
				errs = append(errs, fmt.Errorf("storage[%d].'disk_size' is required", i))
			}
			if storage.DiskControllerIndex >= len(c.DiskControllerType) {
				errs = append(errs, fmt.Errorf("storage[%d].'disk_controller_index' references an unknown disk controller", i))
			}
		}
	}

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
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.VCenterDriver)
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
	for _, disk := range s.Config.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	vm, err := d.CreateVM(&driver.CreateConfig{
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
	_, destroy := state.GetOk("destroy_vm")
	if !cancelled && !halted && !destroy {
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
