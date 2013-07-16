package vmware

import (
	"fmt"
	"runtime"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CompactDisk compacts a virtual disk.
	CompactDisk(string) error

	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string, bool) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// Get the path to the VMware ISO for the given flavor.
	ToolsIsoPath(string) string

	// Get the path to the DHCP leases file for the given device.
	DhcpLeasesPath(string) string

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error.
	Verify() error
}

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver() (Driver, error) {
	var driver Driver

	switch runtime.GOOS {
	case "darwin":
		driver = &Fusion5Driver{
			AppPath: "/Applications/VMware Fusion.app",
		}
	case "linux":
		driver = &Workstation9LinuxDriver{}
	default:
		return nil, fmt.Errorf("can't find driver for OS: %s", runtime.GOOS)
	}

	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
