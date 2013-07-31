// +build windows

package vmware

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func workstationCheckLicense() error {
	// Not implemented on Windows
	return nil
}

func workstationFindVdiskManager() (string, error) {
	path, err := exec.LookPath("vmware-vdiskmanager.exe")
	if err == nil {
		return path, nil
	}

	path, err = workstationVMwareRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "vmware-vdiskmanager.exe"), nil
}

func workstationFindVMware() (string, error) {
	path, err := exec.LookPath("vmware.exe")
	if err == nil {
		return path, nil
	}

	path, err = workstationVMwareRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "vmware.exe"), nil
}

func workstationFindVmrun() (string, error) {
	path, err := exec.LookPath("vmrun.exe")
	if err == nil {
		return path, nil
	}

	path, err = workstationVMwareRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "vmrun.exe"), nil
}

func workstationToolsIsoPath(flavor string) string {
	path, err := workstationVMwareRoot()
	if err != nil {
		return ""
	}

	return filepath.Join(path, flavor+".iso")
}

func workstationDhcpLeasesPath(device string) string {
	programData := os.Getenv("ProgramData")
	if programData == "" {
		return ""
	}

	return filepath.Join(programData, "/VMware/vmnetdhcp.leases")
}

// See http://blog.natefinch.com/2012/11/go-win-stuff.html
//
// This is used by workstationVMwareRoot in order to read some registry data.
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

// This reads the VMware installation path from the Windows registry.
func workstationVMwareRoot() (s string, err error) {
	key := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\vmware.exe`
	subkey := "Path"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	s = strings.Replace(s, "\\", "/", -1)
	return
}
