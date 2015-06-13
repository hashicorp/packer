package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
)

// Workstation9Driver is a driver that can run VMware Workstation 9
type Workstation9Driver struct {
	AppPath          string
	VdiskManagerPath string
	VmrunPath        string

	// SSHConfig are the SSH settings for the Fusion VM
	SSHConfig *SSHConfig
}

func (d *Workstation9Driver) Clone(dst, src string) error {
	return errors.New("Cloning is not supported with VMware WS version 9. Please use VMware WS version 10, or greater.")
}

func (d *Workstation9Driver) CompactDisk(diskPath string) error {
	defragCmd := exec.Command(d.VdiskManagerPath, "-d", diskPath)
	if _, _, err := runAndLog(defragCmd); err != nil {
		return err
	}

	shrinkCmd := exec.Command(d.VdiskManagerPath, "-k", diskPath)
	if _, _, err := runAndLog(shrinkCmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) CreateDisk(output string, size string, type_id string) error {
	cmd := exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", "lsilogic", "-t", type_id, output)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) IsRunning(vmxPath string) (bool, error) {
	vmxPath, err := filepath.Abs(vmxPath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(d.VmrunPath, "-T", "ws", "list")
	stdout, _, err := runAndLog(cmd)
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(stdout, "\n") {
		if line == vmxPath {
			return true, nil
		}
	}

	return false, nil
}

func (d *Workstation9Driver) CommHost(state multistep.StateBag) (string, error) {
	return CommHost(d.SSHConfig)(state)
}

func (d *Workstation9Driver) Start(vmxPath string, headless bool) error {
	guiArgument := "gui"
	if headless {
		guiArgument = "nogui"
	}

	cmd := exec.Command(d.VmrunPath, "-T", "ws", "start", vmxPath, guiArgument)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.VmrunPath, "-T", "ws", "stop", vmxPath, "hard")
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) SuppressMessages(vmxPath string) error {
	return nil
}

func (d *Workstation9Driver) Verify() error {
	var err error
	if d.AppPath == "" {
		if d.AppPath, err = workstationFindVMware(); err != nil {
			return err
		}
	}

	if d.VmrunPath == "" {
		if d.VmrunPath, err = workstationFindVmrun(); err != nil {
			return err
		}
	}

	if d.VdiskManagerPath == "" {
		if d.VdiskManagerPath, err = workstationFindVdiskManager(); err != nil {
			return err
		}
	}

	log.Printf("VMware app path: %s", d.AppPath)
	log.Printf("vmrun path: %s", d.VmrunPath)
	log.Printf("vdisk-manager path: %s", d.VdiskManagerPath)

	if _, err := os.Stat(d.AppPath); err != nil {
		return fmt.Errorf("VMware application not found: %s", d.AppPath)
	}

	if _, err := os.Stat(d.VmrunPath); err != nil {
		return fmt.Errorf("'vmrun' application not found: %s", d.VmrunPath)
	}

	if _, err := os.Stat(d.VdiskManagerPath); err != nil {
		return fmt.Errorf("'vmware-vdiskmanager' application not found: %s", d.VdiskManagerPath)
	}

	// Check to see if it APPEARS to be licensed.
	if err := workstationCheckLicense(); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) ToolsIsoPath(flavor string) string {
	return workstationToolsIsoPath(flavor)
}

func (d *Workstation9Driver) ToolsInstall() error {
	return nil
}

func (d *Workstation9Driver) DhcpLeasesPath(device string) string {
	return workstationDhcpLeasesPath(device)
}

func (d *Workstation9Driver) VmnetnatConfPath() string {
	return workstationVmnetnatConfPath()
}
