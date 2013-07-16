package vmware

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Workstation9LinuxDriver is a driver that can run VMware Workstation 9
// on Linux.
type Workstation9LinuxDriver struct {
	AppPath          string
	VdiskManagerPath string
	VmrunPath        string
}

func (d *Workstation9LinuxDriver) CompactDisk(diskPath string) error {
	defragCmd := exec.Command(d.VdiskManagerPath, "-d", diskPath)
	if _, _, err := d.runAndLog(defragCmd); err != nil {
		return err
	}

	shrinkCmd := exec.Command(d.VdiskManagerPath, "-k", diskPath)
	if _, _, err := d.runAndLog(shrinkCmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9LinuxDriver) CreateDisk(output string, size string) error {
	cmd := exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", "lsilogic", "-t", "1", output)
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9LinuxDriver) IsRunning(vmxPath string) (bool, error) {
	vmxPath, err := filepath.Abs(vmxPath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(d.VmrunPath, "-T", "ws", "list")
	stdout, _, err := d.runAndLog(cmd)
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

func (d *Workstation9LinuxDriver) Start(vmxPath string, headless bool) error {
	guiArgument := "gui"
	if headless {
		guiArgument = "nogui"
	}

	cmd := exec.Command(d.VmrunPath, "-T", "ws", "start", vmxPath, guiArgument)
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9LinuxDriver) Stop(vmxPath string) error {
	cmd := exec.Command(d.VmrunPath, "-T", "ws", "stop", vmxPath, "hard")
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9LinuxDriver) Verify() error {
	if err := d.findApp(); err != nil {
		return fmt.Errorf("VMware Workstation application ('vmware') not found in path.")
	}

	if err := d.findVmrun(); err != nil {
		return fmt.Errorf("Required application 'vmrun' not found in path.")
	}

	if err := d.findVdiskManager(); err != nil {
		return fmt.Errorf("Required application 'vmware-vdiskmanager' not found in path.")
	}

	// Check to see if it APPEARS to be licensed.
	matches, err := filepath.Glob("/etc/vmware/license-*")
	if err != nil {
		return fmt.Errorf("Error looking for VMware license: %s", err)
	}

	if len(matches) == 0 {
		return errors.New("Workstation does not appear to be licensed. Please license it.")
	}

	return nil
}

func (d *Workstation9LinuxDriver) findApp() error {
	path, err := exec.LookPath("vmware")
	if err != nil {
		return err
	}
	d.AppPath = path
	return nil
}

func (d *Workstation9LinuxDriver) findVdiskManager() error {
	path, err := exec.LookPath("vmware-vdiskmanager")
	if err != nil {
		return err
	}
	d.VdiskManagerPath = path
	return nil
}

func (d *Workstation9LinuxDriver) findVmrun() error {
	path, err := exec.LookPath("vmrun")
	if err != nil {
		return err
	}
	d.VmrunPath = path
	return nil
}

func (d *Workstation9LinuxDriver) ToolsIsoPath(flavor string) string {
	return "/usr/lib/vmware/isoimages/" + flavor + ".iso"
}

func (d *Workstation9LinuxDriver) DhcpLeasesPath(device string) string {
	return "/etc/vmware/" + device + "/dhcpd/dhcpd.leases"
}

func (d *Workstation9LinuxDriver) runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %v", cmd.Path, cmd.Args[1:])
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("VMware error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return stdout.String(), stderr.String(), err
}
