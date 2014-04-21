package common

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
)

// Player5LinuxDriver is a driver that can run VMware Player 5 on Linux.
type Player5LinuxDriver struct {
	AppPath          string
	VdiskManagerPath string
	QemuImgPath      string
	VmrunPath        string

	// SSHConfig are the SSH settings for the Fusion VM
	SSHConfig *SSHConfig
}

func (d *Player5LinuxDriver) Clone(dst, src string) error {
	return errors.New("Cloning is not supported with Player 5. Please use Player 6+.")
}

func (d *Player5LinuxDriver) CompactDisk(diskPath string) error {
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

func (d *Player5LinuxDriver) qemuCompactDisk(diskPath string) error {
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

func (d *Player5LinuxDriver) CreateDisk(output string, size string, type_id string) error {
	var cmd *exec.Cmd
	if d.QemuImgPath != "" {
		cmd = exec.Command(d.QemuImgPath, "create", "-f", "vmdk", "-o", "compat6", output, size)
	} else {
		cmd = exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", "lsilogic", "-t", type_id, output)
	}
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player5LinuxDriver) IsRunning(vmxPath string) (bool, error) {
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

func (d *Player5LinuxDriver) SSHAddress(state multistep.StateBag) (string, error) {
	return SSHAddressFunc(d.SSHConfig)(state)
}

func (d *Player5LinuxDriver) Start(vmxPath string, headless bool) error {
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

func (d *Player5LinuxDriver) Stop(vmxPath string) error {
	cmd := exec.Command(d.VmrunPath, "-T", "player", "stop", vmxPath, "hard")
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player5LinuxDriver) SuppressMessages(vmxPath string) error {
	return nil
}

func (d *Player5LinuxDriver) Verify() error {
	if err := d.findApp(); err != nil {
		return fmt.Errorf("VMware Player application ('vmplayer') not found in path.")
	}

	if err := d.findVmrun(); err != nil {
		return fmt.Errorf("Critical application 'vmrun' not found in path.")
	}

	if err := d.findVdiskManager(); err != nil {
		if err := d.findQemuImg(); err != nil {
			return fmt.Errorf(
				"Neither 'vmware-vdiskmanager', nor 'qemu-img' found in path.\n" +
					"One of these is required to configure disks for VMware Player.")
		}
	}

	return nil
}

func (d *Player5LinuxDriver) findApp() error {
	path, err := exec.LookPath("vmplayer")
	if err != nil {
		return err
	}
	d.AppPath = path
	return nil
}

func (d *Player5LinuxDriver) findVdiskManager() error {
	path, err := exec.LookPath("vmware-vdiskmanager")
	if err != nil {
		return err
	}
	d.VdiskManagerPath = path
	return nil
}

func (d *Player5LinuxDriver) findQemuImg() error {
	path, err := exec.LookPath("qemu-img")
	if err != nil {
		return err
	}
	d.QemuImgPath = path
	return nil
}

func (d *Player5LinuxDriver) findVmrun() error {
	path, err := exec.LookPath("vmrun")
	if err != nil {
		return err
	}
	d.VmrunPath = path
	return nil
}

func (d *Player5LinuxDriver) ToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

func (d *Player5LinuxDriver) ToolsInstall() error {
	return nil
}

func (d *Player5LinuxDriver) DhcpLeasesPath(device string) string {
	return "/etc/vmware/" + device + "/dhcpd/dhcpd.leases"
}
