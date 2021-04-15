//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type StorageConfig,DiskConfig

package common

import (
	"fmt"
)

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

type StorageConfig struct {
	// Set VM disk controller type. Example `lsilogic`, `pvscsi`, `nvme`, or `scsi`. Use a list to define additional controllers.
	// Defaults to `lsilogic`. See
	// [SCSI, SATA, and NVMe Storage Controller Conditions, Limitations, and Compatibility](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362.html#GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362)
	// for additional details.
	DiskControllerType []string `mapstructure:"disk_controller_type"`
	// Configures a collection of one or more disks to be provisioned along with the VM. See the [Storage Configuration](#storage-configuration).
	Storage []DiskConfig `mapstructure:"storage"`
}

func (c *StorageConfig) Prepare() []error {
	var errs []error

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

	return errs
}
