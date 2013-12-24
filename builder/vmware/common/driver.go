package common

import (
	"github.com/mitchellh/multistep"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CompactDisk compacts a virtual disk.
	CompactDisk(string) error

	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// SSHAddress returns the SSH address for the VM that is being
	// managed by this driver.
	SSHAddress(multistep.StateBag) (string, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string, bool) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// SuppressMessages modifies the VMX or surrounding directory so that
	// VMware doesn't show any annoying messages.
	SuppressMessages(string) error

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
