// +build windows
// Contributed by Ross Smith II (smithii.com)

package vmware

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Workstation9Driver is a driver that can run VMware Workstation 9
// on Windows.
type Workstation9Driver struct {
	AppPath          string
	VdiskManagerPath string
	VmrunPath        string
}

func (d *Workstation9Driver) CompactDisk(diskPath string) error {
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

func (d *Workstation9Driver) CreateDisk(output string, size string) error {
	cmd := exec.Command(d.VdiskManagerPath, "-c", "-s", size, "-a", "lsilogic", "-t", "1", output)
	if _, _, err := d.runAndLog(cmd); err != nil {
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

func (d *Workstation9Driver) Start(vmxPath string, headless bool) error {
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

func (d *Workstation9Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.VmrunPath, "-T", "ws", "stop", vmxPath, "hard")
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Workstation9Driver) Verify() error {
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
	/*
		matches, err := filepath.Glob("/etc/vmware/license-*")
		if err != nil {
			return fmt.Errorf("Error looking for VMware license: %s", err)
		}

		if len(matches) == 0 {
			return errors.New("Workstation does not appear to be licensed. Please license it.")
		}
	*/
	return nil
}

func (d *Workstation9Driver) findApp() error {
	path, err := exec.LookPath("vmware.exe")
	if err != nil {
		path, err := getVmwarePath()
		if err != nil {
			return err
		}
		path += "vmware.exe"
	}
	path = strings.Replace(path, "\\", "/", -1)
	log.Printf("Using '%s' for vmware path", path)
	d.AppPath = path

	return nil
}

func (d *Workstation9Driver) findVdiskManager() error {
	path, err := exec.LookPath("vmware-vdiskmanager.exe")
	if err != nil {
		path, err := getVmwarePath()
		if err != nil {
			return err
		}
		path += "vmware-vdiskmanager.exe"
	}
	path = strings.Replace(path, "\\", "/", -1)
	log.Printf("Using '%s' for vmware-vdiskmanager path", path)
	d.VdiskManagerPath = path
	return nil
}

func (d *Workstation9Driver) findVmrun() error {
	path, err := exec.LookPath("vmrun.exe")
	if err != nil {
		path, err := getVmwarePath()
		if err != nil {
			return err
		}
		path += "vmrun.exe"
	}
	path = strings.Replace(path, "\\", "/", -1)
	log.Printf("Using '%s' for vmrun path", path)
	d.VmrunPath = path
	return nil
}

func (d *Workstation9Driver) ToolsIsoPath(flavor string) string {
	path, err := getVmwarePath()
	if err != nil {
		return ""
	} else {
		return path + flavor + ".iso"
	}
}

func (d *Workstation9Driver) DhcpLeasesPath(device string) string {
	programData := os.Getenv("ProgramData")
	rv := programData + "/VMware/vmnetdhcp.leases"
	if _, err := os.Stat(rv); os.IsNotExist(err) {
		log.Printf("File not found: '%s' (found '%s' in %%ProgramData%%)", rv, programData)
		return ""
	}
	return rv
}

func (d *Workstation9Driver) runAndLog(cmd *exec.Cmd) (string, string, error) {
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

// see http://blog.natefinch.com/2012/11/go-win-stuff.html

func readRegString(hive syscall.Handle, subKeyPath, valueName string) (value string, err error) {
	var h syscall.Handle
	err = syscall.RegOpenKeyEx(hive, syscall.StringToUTF16Ptr(subKeyPath), 0, syscall.KEY_READ, &h)
	if err != nil {
		return
	}
	defer syscall.RegCloseKey(h)

	var typ uint32
	var bufSize uint32

	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		nil,
		&bufSize)
	if err != nil {
		return
	}

	data := make([]uint16, bufSize/2+1)

	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&data[0])),
		&bufSize)
	if err != nil {
		return
	}

	return syscall.UTF16ToString(data), nil
}

func getVmwarePath() (s string, e error) {
	key := "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\App Paths\\vmware.exe"
	subkey := "Path"
	s, e = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if e != nil {
		log.Printf("Unable to read registry key %s\\%s", key, subkey)
		return "", e
	}
	log.Printf("Found '%s' in registry key %s\\%s", s, key, subkey)
	s = strings.Replace(s, "\\", "/", -1)
	return s, nil
}
