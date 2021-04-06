package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// Workstation9Driver is a driver that can run VMware Workstation 9
type Workstation9Driver struct {
	VmwareDriver

	AppPath          string
	VdiskManagerPath string
	VmrunPath        string

	// SSHConfig are the SSH settings for the Fusion VM
	SSHConfig *SSHConfig
}

func NewWorkstation9Driver(config *SSHConfig) Driver {
	return &Workstation9Driver{
		SSHConfig: config,
	}
}

func (d *Workstation9Driver) Clone(dst, src string, linked bool, snapshot string) error {
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

func (d *Workstation9Driver) CreateDisk(output string, size string, adapter_type string, type_id string) error {
	cmd := exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", adapter_type, "-t", type_id, output)
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

	// Assigning the path callbacks to VmwareDriver
	d.VmwareDriver.DhcpLeasesPath = func(device string) string {
		return workstationDhcpLeasesPath(device)
	}

	d.VmwareDriver.DhcpConfPath = func(device string) string {
		return workstationDhcpConfPath(device)
	}

	d.VmwareDriver.VmnetnatConfPath = func(device string) string {
		return workstationVmnetnatConfPath(device)
	}

	d.VmwareDriver.NetworkMapper = func() (NetworkNameMapper, error) {
		pathNetmap := workstationNetmapConfPath()

		// Check that the file for the networkmapper configuration exists. If there's no
		// error, then the file exists and we can proceed to read the configuration out of it.
		if _, err := os.Stat(pathNetmap); err == nil {
			log.Printf("Located networkmapper configuration file using Workstation: %s", pathNetmap)
			return ReadNetmapConfig(pathNetmap)
		}

		// If we weren't able to find the networkmapper configuration file, then fall back
		// to the networking file which might also be in the configuration directory.
		libpath, _ := workstationVMwareRoot()
		pathNetworking := filepath.Join(libpath, "networking")
		if _, err := os.Stat(pathNetworking); err != nil {
			return nil, fmt.Errorf("Could not determine network mappings from files in path: %s", libpath)
		}

		// We were able to successfully stat the file.. So, now we can open a handle to it.
		log.Printf("Located networking configuration file using Workstation: %s", pathNetworking)
		fd, err := os.Open(pathNetworking)
		if err != nil {
			return nil, err
		}
		defer fd.Close()

		// Then we pass the handle to the networking configuration parser.
		return ReadNetworkingConfig(fd)
	}
	return nil
}

func (d *Workstation9Driver) ToolsIsoPath(flavor string) string {
	return workstationToolsIsoPath(flavor)
}

func (d *Workstation9Driver) ToolsInstall() error {
	return nil
}

func (d *Workstation9Driver) GetVmwareDriver() VmwareDriver {
	return d.VmwareDriver
}
