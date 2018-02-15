// +build !windows

// These functions are compatible with WS 9 and 10 on *NIX
package common

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

func workstationCheckLicense() error {
	matches, err := filepath.Glob("/etc/vmware/license-ws-*")
	if err != nil {
		return fmt.Errorf("Error looking for VMware license: %s", err)
	}

	if len(matches) == 0 {
		return errors.New("Workstation does not appear to be licensed. Please license it.")
	}

	return nil
}

func workstationFindVdiskManager() (string, error) {
	return exec.LookPath("vmware-vdiskmanager")
}

func workstationFindVMware() (string, error) {
	return exec.LookPath("vmware")
}

func workstationFindVmrun() (string, error) {
	return exec.LookPath("vmrun")
}

// return the base path to vmware's config on the host
func workstationVMwareRoot() (s string, err error) {
	return "/etc/vmware", nil
}

func workstationDhcpLeasesPath(device string) string {
	base, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, device, "dhcpd/dhcpd.leases")
}

func workstationDhcpConfPath(device string) string {
	base, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, device, "dhcp/dhcp.conf")
}

func workstationVmnetnatConfPath(device string) string {
	base, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, device, "nat/nat.conf")
}

func workstationNetmapConfPath() string {
	base, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, "netmap.conf")
}

func workstationToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

func workstationVerifyVersion(version string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("The VMware WS version %s driver is only supported on Linux, and Windows, at the moment. Your OS: %s", version, runtime.GOOS)
	}

	//TODO(pmyjavec) there is a better way to find this, how?
	//the default will suffice for now.
	vmxpath := "/usr/lib/vmware/bin/vmware-vmx"

	var stderr bytes.Buffer
	cmd := exec.Command(vmxpath, "-v")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return workstationTestVersion(version, stderr.String())
}

func workstationTestVersion(wanted, versionOutput string) error {
	versionRe := regexp.MustCompile(`(?i)VMware Workstation (\d+)\.`)
	matches := versionRe.FindStringSubmatch(versionOutput)
	if matches == nil {
		return fmt.Errorf(
			"Could not find VMware WS version in output: %s", wanted)
	}
	log.Printf("Detected VMware WS version: %s", matches[1])

	return compareVersions(matches[1], wanted, "Workstation")
}
