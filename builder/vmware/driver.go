package vmware

import (
	"os/exec"
	"path/filepath"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string) error

	// Start starts a VM specified by the path to the VMX given.
	Start(string) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error
}

// Fusion5Driver is a driver that can run VMWare Fusion 5.
type Fusion5Driver struct {
	// This is the path to the "VMware Fusion.app"
	AppPath string
}

func (d *Fusion5Driver) CreateDisk(output string, size string) error {
	vdiskPath := filepath.Join(d.AppPath, "Contents", "Library", "vmware-vdiskmanager")
	cmd := exec.Command(vdiskPath, "-c", "-s", size, "-a", "lsilogic", "-t", "1", output)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) Start(vmxPath string) error {
	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "start", vmxPath, "gui")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "stop", vmxPath, "hard")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) vmrunPath() string {
	return filepath.Join(d.AppPath, "Contents", "Library", "vmrun")
}
