package common

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Driver is the interface that talks to Parallels and performs certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the Parallels builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {
	// Compact a virtual disk image.
	CompactDisk(string) error

	// Adds new CD/DVD drive to the VM and returns name of this device
	DeviceAddCDROM(string, string) (string, error)

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
	ToolsISOPath(string) (string, error)

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of Parallels that is installed.
	Version() (string, error)

	// Send scancodes to the vm using the prltype python script.
	SendKeyScanCodes(string, ...string) error

	// Apply default configuration settings to the virtual machine
	SetDefaultConfiguration(string) error

	// Finds the MAC address of the NIC nic0
	MAC(string) (string, error)

	// Finds the IP address of a VM connected that uses DHCP by its MAC address
	IPAddress(string) (string, error)
}

// NewDriver returns a new driver implementation for this version of Parallels
// Desktop, or an error if the driver couldn't be initialized.
func NewDriver() (Driver, error) {
	var drivers map[string]Driver
	var prlctlPath string
	var prlsrvctlPath string
	var supportedVersions []string
	DHCPLeaseFile := "/Library/Preferences/Parallels/parallels_dhcp_leases"

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

	if prlsrvctlPath == "" {
		var err error
		prlsrvctlPath, err = exec.LookPath("prlsrvctl")
		if err != nil {
			return nil, err
		}
	}

	log.Printf("prlsrvctl path: %s", prlsrvctlPath)

	drivers = map[string]Driver{
		"11": &Parallels11Driver{
			Parallels9Driver: Parallels9Driver{
				PrlctlPath:    prlctlPath,
				PrlsrvctlPath: prlsrvctlPath,
				dhcpLeaseFile: DHCPLeaseFile,
			},
		},
		"10": &Parallels10Driver{
			Parallels9Driver: Parallels9Driver{
				PrlctlPath:    prlctlPath,
				PrlsrvctlPath: prlsrvctlPath,
				dhcpLeaseFile: DHCPLeaseFile,
			},
		},
		"9": &Parallels9Driver{
			PrlctlPath:    prlctlPath,
			PrlsrvctlPath: prlsrvctlPath,
			dhcpLeaseFile: DHCPLeaseFile,
		},
	}

	for v, d := range drivers {
		version, _ := d.Version()
		if strings.HasPrefix(version, v) {
			if err := d.Verify(); err != nil {
				return nil, err
			}
			return d, nil
		}
		supportedVersions = append(supportedVersions, v)
	}

	latestDriver := 11
	version, _ := drivers[strconv.Itoa(latestDriver)].Version()
	majVer, _ := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	if majVer > latestDriver {
		log.Printf("Your version of Parallels Desktop for Mac is %s, Packer will use driver for version %d.", version, latestDriver)
		return drivers[strconv.Itoa(latestDriver)], nil
	}

	return nil, fmt.Errorf(
		"Unable to initialize any driver. Supported Parallels Desktop versions: "+
			"%s\n", strings.Join(supportedVersions, ", "))
}
