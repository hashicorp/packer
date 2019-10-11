// +build !windows

// These functions are compatible with WS 9 and 10 on *NIX
package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

func playerFindVdiskManager() (string, error) {
	return exec.LookPath("vmware-vdiskmanager")
}

func playerFindQemuImg() (string, error) {
	return exec.LookPath("qemu-img")
}

func playerFindVMware() (string, error) {
	return exec.LookPath("vmplayer")
}

func playerFindVmrun() (string, error) {
	return exec.LookPath("vmrun")
}

func playerToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

// return the base path to vmware's config on the host
func playerVMwareRoot() (s string, err error) {
	return "/etc/vmware", nil
}

func playerDhcpLeasesPath(device string) string {
	base, err := playerVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}

	// Build the base path to VMware configuration for specified device: `/etc/vmware/${device}`
	devicebase := filepath.Join(base, device)

	// Walk through a list of paths searching for the correct permutation...
	// ...as it appears that in >= WS14 and < WS14, the leases file may be labelled differently.

	// Docs say we should expect: dhcpd/dhcpd.leases
	paths := []string{"dhcpd/dhcpd.leases", "dhcpd/dhcp.leases", "dhcp/dhcpd.leases", "dhcp/dhcp.leases"}
	for _, p := range paths {
		fp := filepath.Join(devicebase, p)
		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			return fp
		}
	}

	log.Printf("Error finding VMWare DHCP Server Leases (dhcpd.leases) under device path: %s", devicebase)
	return ""
}

func playerVmDhcpConfPath(device string) string {
	base, err := playerVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}

	// Build the base path to VMware configuration for specified device: `/etc/vmware/${device}`
	devicebase := filepath.Join(base, device)

	// Walk through a list of paths searching for the correct permutation...
	// ...as it appears that in >= WS14 and < WS14, the dhcp config may be labelled differently.

	// Docs say we should expect: dhcp/dhcp.conf
	paths := []string{"dhcp/dhcp.conf", "dhcp/dhcpd.conf", "dhcpd/dhcp.conf", "dhcpd/dhcpd.conf"}
	for _, p := range paths {
		fp := filepath.Join(devicebase, p)
		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			return fp
		}
	}

	log.Printf("Error finding VMWare DHCP Server Configuration (dhcp.conf) under device path: %s", devicebase)
	return ""
}

func playerVmnetnatConfPath(device string) string {
	base, err := playerVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, device, "nat/nat.conf")
}

func playerNetmapConfPath() string {
	base, err := playerVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
		return ""
	}
	return filepath.Join(base, "netmap.conf")
}

func playerVerifyVersion(version string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("The VMWare Player version %s driver is only supported on Linux, and Windows, at the moment. Your OS: %s", version, runtime.GOOS)
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

	versionRe := regexp.MustCompile(`(?i)VMware Player (\d+)\.`)
	matches := versionRe.FindStringSubmatch(stderr.String())
	if matches == nil {
		return fmt.Errorf(
			"Could not find VMWare Player version in output: %s", stderr.String())
	}
	log.Printf("Detected VMWare Player version: %s", matches[1])

	return compareVersions(matches[1], version, "Player")
}
