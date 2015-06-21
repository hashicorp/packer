package common

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// A driver is able to talk to Parallels and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the Parallels builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {
	// Compact a virtual disk image.
	CompactDisk(string) error

	// Adds new CD/DVD drive to the VM and returns name of this device
	DeviceAddCdRom(string, string) (string, error)

	// Get path to the first virtual disk image
	DiskPath(string) (string, error)

	// Import a VM
	Import(string, string, string, bool) error

	// Checks if the VM with the given name is running.
	IsRunning(string) (bool, error)

	// Stop stops a running machine, forcefully.
	Stop(string) error

	// Prlctl executes the given Prlctl command
	Prlctl(...string) error

	// Get the path to the Parallels Tools ISO for the given flavor.
	ToolsIsoPath(string) (string, error)

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of Parallels that is installed.
	Version() (string, error)

	// Send scancodes to the vm using the prltype python script.
	SendKeyScanCodes(string, ...string) error

	// Apply default сonfiguration settings to the virtual machine
	SetDefaultConfiguration(string) error

	// Finds the MAC address of the NIC nic0
	Mac(string) (string, error)

	// Finds the IP address of a VM connected that uses DHCP by its MAC address
	IpAddress(string) (string, error)
}

func NewDriver() (Driver, error) {
	var drivers map[string]Driver
	var prlctlPath string
	var supportedVersions []string
	dhcp_lease_file := "/Library/Preferences/Parallels/parallels_dhcp_leases"

	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf(
			"Parallels builder works only on \"darwin\" platform!")
	}

	if prlctlPath == "" {
		var err error
		prlctlPath, err = exec.LookPath("prlctl")
		if err != nil {
			return nil, err
		}
	}

	log.Printf("prlctl path: %s", prlctlPath)

	drivers = map[string]Driver{
		"11": &Parallels10Driver{
			Parallels9Driver: Parallels9Driver{
				PrlctlPath:      prlctlPath,
				dhcp_lease_file: dhcp_lease_file,
			},
		},
		"10": &Parallels10Driver{
			Parallels9Driver: Parallels9Driver{
				PrlctlPath:      prlctlPath,
				dhcp_lease_file: dhcp_lease_file,
			},
		},
		"9": &Parallels9Driver{
			PrlctlPath:      prlctlPath,
			dhcp_lease_file: dhcp_lease_file,
		},
	}

	for v, d := range drivers {
		version, _ := d.Version()
		if strings.HasPrefix(version, v) {
			return d, nil
		}
		supportedVersions = append(supportedVersions, v)
	}

	return nil, fmt.Errorf(
		"Unable to initialize any driver. Supported Parallels Desktop versions: "+
			"%s\n", strings.Join(supportedVersions, ", "))
}
