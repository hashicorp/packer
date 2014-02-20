// +build !windows

package common

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

func workstationCheckLicense() error {
	matches, err := filepath.Glob("/etc/vmware/license-*")
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

func workstationDhcpLeasesPath(device string) string {
	return "/etc/vmware/" + device + "/dhcpd/dhcpd.leases"
}

func workstationToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

func workstationVmnetnatConfPath() string {
	return ""
}
