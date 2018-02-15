package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const VMWARE_FUSION_VERSION = "6"

// Fusion6Driver is a driver that can run VMware Fusion 6.
type Fusion6Driver struct {
	Fusion5Driver
}

func (d *Fusion6Driver) Clone(dst, src string) error {
	cmd := exec.Command(d.vmrunPath(),
		"-T", "fusion",
		"clone", src, dst,
		"full")
	if _, _, err := runAndLog(cmd); err != nil {
		if strings.Contains(err.Error(), "parameters was invalid") {
			return fmt.Errorf(
				"Clone is not supported with your version of Fusion. Packer "+
					"only works with Fusion %s Professional or above. Please verify your version.", VMWARE_FUSION_VERSION)
		}

		return err
	}

	return nil
}

func (d *Fusion6Driver) Verify() error {
	if err := d.Fusion5Driver.Verify(); err != nil {
		return err
	}

	vmxpath := filepath.Join(d.AppPath, "Contents", "Library", "vmware-vmx")
	if _, err := os.Stat(vmxpath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vmware-vmx could not be found at path: %s",
				vmxpath)
		}

		return err
	}

	var stderr bytes.Buffer
	cmd := exec.Command(vmxpath, "-v")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	versionRe := regexp.MustCompile(`(?i)VMware [a-z0-9-]+ (\d+)\.`)
	matches := versionRe.FindStringSubmatch(stderr.String())
	if matches == nil {
		return fmt.Errorf(
			"Couldn't find VMware version in output: %s", stderr.String())
	}
	log.Printf("Detected VMware version: %s", matches[1])

	libpath := filepath.Join("/", "Library", "Preferences", "VMware Fusion")

	d.VmwareDriver.DhcpLeasesPath = func(device string) string {
		return "/var/db/vmware/vmnet-dhcpd-" + device + ".leases"
	}
	d.VmwareDriver.DhcpConfPath = func(device string) string {
		return filepath.Join(libpath, device, "dhcpd.conf")
	}

	d.VmwareDriver.VmnetnatConfPath = func(device string) string {
		return filepath.Join(libpath, device, "nat.conf")
	}
	d.VmwareDriver.NetworkMapper = func() (NetworkNameMapper, error) {
		pathNetworking := filepath.Join(libpath, "networking")
		if _, err := os.Stat(pathNetworking); err != nil {
			return nil, fmt.Errorf("Could not find networking conf file: %s", pathNetworking)
		}

		fd, err := os.Open(pathNetworking)
		if err != nil {
			return nil, err
		}
		defer fd.Close()

		return ReadNetworkingConfig(fd)
	}

	return compareVersions(matches[1], VMWARE_FUSION_VERSION, "Fusion Professional")
}

func (d *Fusion6Driver) GetVmwareDriver() VmwareDriver {
	return d.Fusion5Driver.VmwareDriver
}
