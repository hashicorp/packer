package common

import (
	"log"
	"os/exec"
)

// A driver is able to talk to Parallels and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the Parallels builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {
	// Import a VM
	Import(string, string, string) error

	// Checks if the VM with the given name is running.
	IsRunning(string) (bool, error)

	// Stop stops a running machine, forcefully.
	Stop(string) error

	// Prlctl executes the given Prlctl command
	Prlctl(...string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of Parallels that is installed.
	Version() (string, error)

	// Send scancodes to the vm using the prltype tool.
	SendKeyScanCodes(string, ...string) error

	// Finds the MAC address of the NIC nic0
	Mac(string) (string, error)

	// Finds the IP address of a VM connected that uses DHCP by its MAC address
	IpAddress(string) (string, error)
}

func NewDriver() (Driver, error) {
	var prlctlPath string

	if prlctlPath == "" {
		var err error
		prlctlPath, err = exec.LookPath("prlctl")
		if err != nil {
			return nil, err
		}
	}

	log.Printf("prlctl path: %s", prlctlPath)
	driver := &Parallels9Driver{prlctlPath}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
