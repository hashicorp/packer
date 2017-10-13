package common

import (
	"fmt"
	"log"

	powershell "github.com/hashicorp/packer/common/powershell"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	DefaultDiskSize = 40 * 1024        // ~40GB
	MinDiskSize     = 256              // 256MB
	MaxDiskSize     = 64 * 1024 * 1024 // 64TB

	DefaultRamSize                 = 1 * 1024  // 1GB
	MinRamSize                     = 32        // 32MB
	MaxRamSize                     = 32 * 1024 // 32GB
	MinNestedVirtualizationRamSize = 4 * 1024  // 4GB

	LowRam = 256 // 256MB
)

type SizeConfig struct {
	// The size, in megabytes, of the hard disk to create for the VM.
	// By default, this is 130048 (about 127 GB).
	DiskSize uint `mapstructure:"disk_size"`
	// The size, in megabytes, of the computer memory in the VM.
	// By default, this is 1024 (about 1 GB).
	RamSize uint `mapstructure:"ram_size"`
}

func (c *SizeConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.checkDiskSize(); err != nil {
		errs = append(errs, err)
	}
	if err := c.checkRamSize(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (c *SizeConfig) ValidateAvailable() string {
	if powershellAvailable, _, _ := powershell.IsPowershellAvailable(); powershellAvailable {
		freeMB := powershell.GetHostAvailableMemory()

		if (freeMB - float64(c.RamSize)) < LowRam {
			return "Hyper-V might fail to create a VM if there is not enough free memory in the system."
		}
	}

	return ""
}

func (c *SizeConfig) ValidateMinimum() string {
	if c.RamSize < MinNestedVirtualizationRamSize {
		return "For nested virtualization, when virtualization extension is enabled, there should be 4GB or more memory set for the vm, otherwise Hyper-V may fail to start any nested VMs."
	}
	return ""
}

func (c *SizeConfig) checkDiskSize() error {
	if c.DiskSize == 0 {
		c.DiskSize = DefaultDiskSize
	}

	log.Println(fmt.Sprintf("%s: %v", "DiskSize", c.DiskSize))

	if c.DiskSize < MinDiskSize {
		return fmt.Errorf("disk_size: Virtual machine requires disk space >= %v GB, but defined: %v", MinDiskSize, c.DiskSize/1024)
	} else if c.DiskSize > MaxDiskSize {
		return fmt.Errorf("disk_size: Virtual machine requires disk space <= %v GB, but defined: %v", MaxDiskSize, c.DiskSize/1024)
	}

	return nil
}

func (c *SizeConfig) checkRamSize() error {
	if c.RamSize == 0 {
		c.RamSize = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", c.RamSize))

	if c.RamSize < MinRamSize {
		return fmt.Errorf("ram_size: Virtual machine requires memory size >= %v MB, but defined: %v", MinRamSize, c.RamSize)
	} else if c.RamSize > MaxRamSize {
		return fmt.Errorf("ram_size: Virtual machine requires memory size <= %v MB, but defined: %v", MaxRamSize, c.RamSize)
	}

	return nil
}
