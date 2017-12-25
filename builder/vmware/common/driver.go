package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// Clone clones the VMX and the disk to the destination path. The
	// destination is a path to the VMX file. The disk will be copied
	// to that same directory.
	Clone(dst string, src string) error

	// CompactDisk compacts a virtual disk.
	CompactDisk(string) error

	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string, string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string, bool) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// SuppressMessages modifies the VMX or surrounding directory so that
	// VMware doesn't show any annoying messages.
	SuppressMessages(string) error

	// Get the path to the VMware ISO for the given flavor.
	ToolsIsoPath(string) string

	// Attach the VMware tools ISO
	ToolsInstall() error

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error. Each vmware
	// driver should assign the VmwareMachine callback functions for locating
	// paths within this function.
	Verify() error

	/// This is to establish a connection to the guest
	CommHost(multistep.StateBag) (string, error)

	/// These methods are generally implemented by the VmwareDriver
	/// structure within this file. A driver implementation can
	/// reimplement these, though, if it wants.
	GetVmwareDriver() VmwareDriver

	// Get the guest hw address for the vm
	GuestAddress(multistep.StateBag) (string, error)

	// Get the guest ip address for the vm
	GuestIP(multistep.StateBag) (string, error)

	// Get the host hw address for the vm
	HostAddress(multistep.StateBag) (string, error)

	// Get the host ip address for the vm
	HostIP(multistep.StateBag) (string, error)
}

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver(dconfig *DriverConfig, config *SSHConfig) (Driver, error) {
	drivers := []Driver{}

	switch runtime.GOOS {
	case "darwin":
		drivers = []Driver{
			&Fusion6Driver{
				Fusion5Driver: Fusion5Driver{
					AppPath:   dconfig.FusionAppPath,
					SSHConfig: config,
				},
			},
			&Fusion5Driver{
				AppPath:   dconfig.FusionAppPath,
				SSHConfig: config,
			},
		}
	case "linux":
		fallthrough
	case "windows":
		drivers = []Driver{
			&Workstation10Driver{
				Workstation9Driver: Workstation9Driver{
					SSHConfig: config,
				},
			},
			&Workstation9Driver{
				SSHConfig: config,
			},
			&Player6Driver{
				Player5Driver: Player5Driver{
					SSHConfig: config,
				},
			},
			&Player5Driver{
				SSHConfig: config,
			},
		}
	default:
		return nil, fmt.Errorf("can't find driver for OS: %s", runtime.GOOS)
	}

	errs := ""
	for _, driver := range drivers {
		err := driver.Verify()
		log.Printf("Testing vmware driver %T. Success: %t",
			driver, err == nil)

		if err == nil {
			return driver, nil
		}
		errs += "* " + err.Error() + "\n"
	}

	return nil, fmt.Errorf(
		"Unable to initialize any driver for this platform. The errors\n"+
			"from each driver are shown below. Please fix at least one driver\n"+
			"to continue:\n%s", errs)
}

func runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		message := stderrString
		if message == "" {
			message = stdoutString
		}

		err = fmt.Errorf("VMware error: %s", message)

		// If "unknown error" is in there, add some additional notes
		re := regexp.MustCompile(`(?i)unknown error`)
		if re.MatchString(message) {
			err = fmt.Errorf(
				"%s\n\n%s", err,
				"Packer detected a VMware 'Unknown Error'. Unfortunately VMware\n"+
					"often has extremely vague error messages such as this and Packer\n"+
					"itself can't do much about that. Please check the vmware.log files\n"+
					"created by VMware when a VM is started (in the directory of the\n"+
					"vmx file), which often contains more detailed error information.")
		}
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	// Replace these for Windows, we only want to deal with Unix
	// style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}

func normalizeVersion(version string) (string, error) {
	i, err := strconv.Atoi(version)
	if err != nil {
		return "", fmt.Errorf(
			"VMware version '%s' is not numeric", version)
	}

	return fmt.Sprintf("%02d", i), nil
}

func compareVersions(versionFound string, versionWanted string, product string) error {
	found, err := normalizeVersion(versionFound)
	if err != nil {
		return err
	}

	wanted, err := normalizeVersion(versionWanted)
	if err != nil {
		return err
	}

	if found < wanted {
		return fmt.Errorf(
			"VMware %s version %s, or greater, is required. Found version: %s", product, versionWanted, versionFound)
	}

	return nil
}

/// helper functions that read configuration information from a file
// read the network<->device configuration out of the specified path
func ReadNetmapConfig(path string) (NetworkMap, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return ReadNetworkMap(fd)
}

// read the dhcp configuration out of the specified path
func ReadDhcpConfig(path string) (DhcpConfiguration, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return ReadDhcpConfiguration(fd)
}

// read the VMX configuration from the specified path
func readVMXConfig(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return map[string]string{}, err
	}
	defer f.Close()

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return map[string]string{}, err
	}
	return ParseVMX(string(vmxBytes)), nil
}

// read the connection type out of a vmx configuration
func readCustomDeviceName(vmxData map[string]string) (string, error) {

	connectionType, ok := vmxData["ethernet0.connectiontype"]
	if !ok || connectionType != "custom" {
		return "", fmt.Errorf("Unable to determine the device name for the connection type : %s", connectionType)
	}

	device, ok := vmxData["ethernet0.vnet"]
	if !ok || device == "" {
		return "", fmt.Errorf("Unable to determine the device name for the connection type \"%s\" : %s", connectionType, device)
	}
	return device, nil
}

// This VmwareDriver is a base class that contains default methods
// that a Driver can use or implement themselves.
type VmwareDriver struct {
	/// These methods define paths that are utilized by the driver
	/// A driver must overload these in order to point to the correct
	/// files so that the address detection (ip and ethernet) machinery
	/// works.
	DhcpLeasesPath   func(string) string
	DhcpConfPath     func(string) string
	VmnetnatConfPath func(string) string

	/// This method returns an object with the NetworkNameMapper interface
	/// that maps network to device and vice-versa.
	NetworkMapper func() (NetworkNameMapper, error)
}

func (d *VmwareDriver) GuestAddress(state multistep.StateBag) (string, error) {
	vmxPath := state.Get("vmx_path").(string)

	log.Println("Lookup up IP information...")
	vmxData, err := readVMXConfig(vmxPath)
	if err != nil {
		return "", err
	}

	var ok bool
	macAddress := ""
	if macAddress, ok = vmxData["ethernet0.address"]; !ok || macAddress == "" {
		if macAddress, ok = vmxData["ethernet0.generatedaddress"]; !ok || macAddress == "" {
			return "", errors.New("couldn't find MAC address in VMX")
		}
	}

	res, err := net.ParseMAC(macAddress)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func (d *VmwareDriver) GuestIP(state multistep.StateBag) (string, error) {

	// grab network mapper
	netmap, err := d.NetworkMapper()
	if err != nil {
		return "", err
	}

	// convert the stashed network to a device
	network := state.Get("vmnetwork").(string)
	device, err := netmap.NameIntoDevice(network)

	// we were unable to find the device, maybe it's a custom one...
	// so, check to see if it's in the .vmx configuration
	if err != nil || network == "custom" {
		vmxPath := state.Get("vmx_path").(string)
		vmxData, err := readVMXConfig(vmxPath)
		if err != nil {
			return "", err
		}

		device, err = readCustomDeviceName(vmxData)
		if err != nil {
			return "", err
		}
	}

	// figure out our MAC address for looking up the guest address
	MACAddress, err := d.GuestAddress(state)
	if err != nil {
		return "", err
	}

	// figure out the correct dhcp leases
	dhcpLeasesPath := d.DhcpLeasesPath(device)
	log.Printf("DHCP leases path: %s", dhcpLeasesPath)
	if dhcpLeasesPath == "" {
		return "", errors.New("no DHCP leases path found.")
	}

	// open up the lease and read its contents
	fh, err := os.Open(dhcpLeasesPath)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	dhcpBytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return "", err
	}

	// start grepping through the file looking for fields that we care about
	var lastIp string
	var lastLeaseEnd time.Time

	var curIp string
	var curLeaseEnd time.Time

	ipLineRe := regexp.MustCompile(`^lease (.+?) {$`)
	endTimeLineRe := regexp.MustCompile(`^\s*ends \d (.+?);$`)
	macLineRe := regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)

	for _, line := range strings.Split(string(dhcpBytes), "\n") {
		// Need to trim off CR character when running in windows
		line = strings.TrimRight(line, "\r")

		matches := ipLineRe.FindStringSubmatch(line)
		if matches != nil {
			lastIp = matches[1]
			continue
		}

		matches = endTimeLineRe.FindStringSubmatch(line)
		if matches != nil {
			lastLeaseEnd, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		// If the mac address matches and this lease ends farther in the
		// future than the last match we might have, then choose it.
		matches = macLineRe.FindStringSubmatch(line)
		if matches != nil && strings.EqualFold(matches[1], MACAddress) && curLeaseEnd.Before(lastLeaseEnd) {
			curIp = lastIp
			curLeaseEnd = lastLeaseEnd
		}
	}
	if curIp == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases at %s", MACAddress, dhcpLeasesPath)
	}
	return curIp, nil
}

func (d *VmwareDriver) HostAddress(state multistep.StateBag) (string, error) {

	// grab mapper for converting network<->device
	netmap, err := d.NetworkMapper()
	if err != nil {
		return "", err
	}

	// convert network to name
	network := state.Get("vmnetwork").(string)
	device, err := netmap.NameIntoDevice(network)

	// we were unable to find the device, maybe it's a custom one...
	// so, check to see if it's in the .vmx configuration
	if err != nil || network == "custom" {
		vmxPath := state.Get("vmx_path").(string)
		vmxData, err := readVMXConfig(vmxPath)
		if err != nil {
			return "", err
		}

		device, err = readCustomDeviceName(vmxData)
		if err != nil {
			return "", err
		}
	}

	// parse dhcpd configuration
	pathDhcpConfig := d.DhcpConfPath(device)
	if _, err := os.Stat(pathDhcpConfig); err != nil {
		return "", fmt.Errorf("Could not find vmnetdhcp conf file: %s", pathDhcpConfig)
	}

	config, err := ReadDhcpConfig(pathDhcpConfig)
	if err != nil {
		return "", err
	}

	// find the entry configured in the dhcpd
	interfaceConfig, err := config.HostByName(device)
	if err != nil {
		return "", err
	}

	// finally grab the hardware address
	address, err := interfaceConfig.Hardware()
	if err == nil {
		return address.String(), nil
	}

	// we didn't find it, so search through our interfaces for the device name
	interfaceList, err := net.Interfaces()
	if err == nil {
		return "", err
	}

	names := make([]string, 0)
	for _, intf := range interfaceList {
		if strings.HasSuffix(strings.ToLower(intf.Name), device) {
			return intf.HardwareAddr.String(), nil
		}
		names = append(names, intf.Name)
	}
	return "", fmt.Errorf("Unable to find device %s : %v", device, names)
}

func (d *VmwareDriver) HostIP(state multistep.StateBag) (string, error) {

	// grab mapper for converting network<->device
	netmap, err := d.NetworkMapper()
	if err != nil {
		return "", err
	}

	// convert network to name
	network := state.Get("vmnetwork").(string)
	device, err := netmap.NameIntoDevice(network)

	// we were unable to find the device, maybe it's a custom one...
	// so, check to see if it's in the .vmx configuration
	if err != nil || network == "custom" {
		vmxPath := state.Get("vmx_path").(string)
		vmxData, err := readVMXConfig(vmxPath)
		if err != nil {
			return "", err
		}

		device, err = readCustomDeviceName(vmxData)
		if err != nil {
			return "", err
		}
	}

	// parse dhcpd configuration
	pathDhcpConfig := d.DhcpConfPath(device)
	if _, err := os.Stat(pathDhcpConfig); err != nil {
		return "", fmt.Errorf("Could not find vmnetdhcp conf file: %s", pathDhcpConfig)
	}
	config, err := ReadDhcpConfig(pathDhcpConfig)
	if err != nil {
		return "", err
	}

	// find the entry configured in the dhcpd
	interfaceConfig, err := config.HostByName(device)
	if err != nil {
		return "", err
	}

	address, err := interfaceConfig.IP4()
	if err != nil {
		return "", err
	}

	return address.String(), nil
}
