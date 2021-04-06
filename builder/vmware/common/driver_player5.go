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

// Player5Driver is a driver that can run VMware Player 5 on Linux.
type Player5Driver struct {
	VmwareDriver

	AppPath          string
	VdiskManagerPath string
	QemuImgPath      string
	VmrunPath        string

	// SSHConfig are the SSH settings for the Fusion VM
	SSHConfig *SSHConfig
}

func NewPlayer5Driver(config *SSHConfig) Driver {
	return &Player5Driver{
		SSHConfig: config,
	}
}

func (d *Player5Driver) Clone(dst, src string, linked bool, snapshot string) error {
	return errors.New("Cloning is not supported with VMWare Player version 5. Please use VMWare Player version 6, or greater.")
}

func (d *Player5Driver) CompactDisk(diskPath string) error {
	if d.QemuImgPath != "" {
		return d.qemuCompactDisk(diskPath)
	}

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

func (d *Player5Driver) qemuCompactDisk(diskPath string) error {
	cmd := exec.Command(d.QemuImgPath, "convert", "-f", "vmdk", "-O", "vmdk", "-o", "compat6", diskPath, diskPath+".new")
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	if err := os.Remove(diskPath); err != nil {
		return err
	}

	if err := os.Rename(diskPath+".new", diskPath); err != nil {
		return err
	}

	return nil
}

func (d *Player5Driver) CreateDisk(output string, size string, adapter_type string, type_id string) error {
	var cmd *exec.Cmd
	if d.QemuImgPath != "" {
		cmd = exec.Command(d.QemuImgPath, "create", "-f", "vmdk", "-o", "compat6", output, size)
	} else {
		cmd = exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", adapter_type, "-t", type_id, output)
	}
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player5Driver) IsRunning(vmxPath string) (bool, error) {
	vmxPath, err := filepath.Abs(vmxPath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(d.VmrunPath, "-T", "player", "list")
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

func (d *Player5Driver) CommHost(state multistep.StateBag) (string, error) {
	return CommHost(d.SSHConfig)(state)
}

func (d *Player5Driver) Start(vmxPath string, headless bool) error {
	guiArgument := "gui"
	if headless {
		guiArgument = "nogui"
	}

	cmd := exec.Command(d.VmrunPath, "-T", "player", "start", vmxPath, guiArgument)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player5Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.VmrunPath, "-T", "player", "stop", vmxPath, "hard")
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player5Driver) SuppressMessages(vmxPath string) error {
	return nil
}

func (d *Player5Driver) Verify() error {
	var err error
	if d.AppPath == "" {
		if d.AppPath, err = playerFindVMware(); err != nil {
			return err
		}
	}

	if d.VmrunPath == "" {
		if d.VmrunPath, err = playerFindVmrun(); err != nil {
			return err
		}
	}

	if d.VdiskManagerPath == "" {
		d.VdiskManagerPath, err = playerFindVdiskManager()
	}

	if d.VdiskManagerPath == "" && d.QemuImgPath == "" {
		d.QemuImgPath, err = playerFindQemuImg()
	}

	if err != nil {
		return fmt.Errorf(
			"Neither 'vmware-vdiskmanager', nor 'qemu-img' found in path.\n" +
				"One of these is required to configure disks for VMware Player.")
	}

	log.Printf("VMware app path: %s", d.AppPath)
	log.Printf("vmrun path: %s", d.VmrunPath)
	log.Printf("vdisk-manager path: %s", d.VdiskManagerPath)
	log.Printf("qemu-img path: %s", d.QemuImgPath)

	if _, err := os.Stat(d.AppPath); err != nil {
		return fmt.Errorf("VMware application not found: %s", d.AppPath)
	}

	if _, err := os.Stat(d.VmrunPath); err != nil {
		return fmt.Errorf("'vmrun' application not found: %s", d.VmrunPath)
	}

	if d.VdiskManagerPath != "" {
		_, err = os.Stat(d.VdiskManagerPath)
	} else {
		_, err = os.Stat(d.QemuImgPath)
	}

	if err != nil {
		return fmt.Errorf(
			"Neither 'vmware-vdiskmanager', nor 'qemu-img' found in path.\n" +
				"One of these is required to configure disks for VMware Player.")
	}

	// Assigning the path callbacks to VmwareDriver
	d.VmwareDriver.DhcpLeasesPath = func(device string) string {
		return playerDhcpLeasesPath(device)
	}

	d.VmwareDriver.DhcpConfPath = func(device string) string {
		return playerVmDhcpConfPath(device)
	}

	d.VmwareDriver.VmnetnatConfPath = func(device string) string {
		return playerVmnetnatConfPath(device)
	}

	d.VmwareDriver.NetworkMapper = func() (NetworkNameMapper, error) {
		pathNetmap := playerNetmapConfPath()

		// If we were able to find the file (no error), then we can proceed with reading
		// the networkmapper configuration.
		if _, err := os.Stat(pathNetmap); err == nil {
			log.Printf("Located networkmapper configuration file using Player: %s", pathNetmap)
			return ReadNetmapConfig(pathNetmap)
		}

		// If we weren't able to find the networkmapper configuration file, then fall back
		// to the networking file which might also be in the configuration directory.
		libpath, _ := playerVMwareRoot()
		pathNetworking := filepath.Join(libpath, "networking")
		if _, err := os.Stat(pathNetworking); err != nil {
			return nil, fmt.Errorf("Could not determine network mappings from files in path: %s", libpath)
		}

		// We were able to successfully stat the file.. So, now we can open a handle to it.
		log.Printf("Located networking configuration file using Player: %s", pathNetworking)
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

func (d *Player5Driver) ToolsIsoPath(flavor string) string {
	return playerToolsIsoPath(flavor)
}

func (d *Player5Driver) ToolsInstall() error {
	return nil
}

func (d *Player5Driver) GetVmwareDriver() VmwareDriver {
	return d.VmwareDriver
}
