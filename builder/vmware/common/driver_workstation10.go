package common

import (
	"os/exec"
)

const VMWARE_WS_VERSION = "10"

// Workstation10Driver is a driver that can run VMware Workstation 10
// installations.

type Workstation10Driver struct {
	Workstation9Driver
}

func NewWorkstation10Driver(config *SSHConfig) Driver {
	return &Workstation10Driver{
		Workstation9Driver: Workstation9Driver{
			SSHConfig: config,
		},
	}
}

func (d *Workstation10Driver) Clone(dst, src string, linked bool) error {

	var cloneType string
	if linked {
		cloneType = "linked"
	} else {
		cloneType = "full"
	}

	cmd := exec.Command(d.Workstation9Driver.VmrunPath,
		"-T", "ws",
		"clone", src, dst,
		cloneType)

	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation10Driver) Verify() error {
	if err := d.Workstation9Driver.Verify(); err != nil {
		return err
	}

	return workstationVerifyVersion(VMWARE_WS_VERSION)
}

func (d *Workstation10Driver) GetVmwareDriver() VmwareDriver {
	return d.Workstation9Driver.VmwareDriver
}
