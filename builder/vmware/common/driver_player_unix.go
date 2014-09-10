// +build !windows

// These functions are compatible with WS 9 and 10 on *NIX
package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
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

func playerDhcpLeasesPath(device string) string {
	return "/etc/vmware/" + device + "/dhcpd/dhcpd.leases"
}

func playerToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

func playerVmnetnatConfPath() string {
	return ""
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
