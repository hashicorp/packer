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
	path, _ := exec.LookPath("vmware-vdiskmanager.exe")
	if fileExists(path) {
		return path, nil
	}

	return findProgramFile("vmware-vdiskmanager.exe"), nil
}

func workstationFindVMware() (string, error) {
	path, _ := exec.LookPath("vmware.exe")
	if fileExists(path) {
		return path, nil
	}

	return findProgramFile("vmware.exe"), nil
}

func workstationFindVmrun() (string, error) {
	path, _ := exec.LookPath("vmrun.exe")
	if fileExists(path) {
		return path, nil
	}

	return findProgramFile("vmrun.exe"), nil
}

func workstationToolsIsoPath(flavor string) string {
	return findProgramFile(flavor + ".iso")
}

func workstationDhcpLeasesPath(device string) string {
	path, _ := workstationVmnetDhcpLeasesPathFromRegistry()

	if fileExists(path) {
		return path
	}

	return findDataFile("vmnetdhcp.leases")
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

	return normalizePath(s), nil
}

// This reads the VMware DHCP leases path from the Windows registry.
func workstationVmnetDhcpLeasesPathFromRegistry() (s string, err error) {
	key := "SYSTEM\\CurrentControlSet\\services\\VMnetDHCP\\Parameters"
	subkey := "LeaseFile"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

func fileExists(file string) bool {
	if file == "" {
		return false
	}

	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Println(err.Error())
	}

	return true
}

func normalizePath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.TrimRight(path, "/")
	return path
}

type Paths [][]string

func findFile(file string, paths Paths) string {
	for _, a := range paths {
		if a[0] == "" {
			continue
		}

		path := filepath.Join(a[0], a[1], file)

		path = normalizePath(path)

		log.Printf("Searching for file '%s'", path)

		if fileExists(path) {
			log.Printf("Found file '%s'", path)
			return path
		}
	}

	log.Printf("File not found: '%s'", file)

	return ""
}

func findProgramFile(file string) string {
	path, _ := workstationVMwareRoot()

	paths := Paths{
		[]string{os.Getenv("VMWARE_HOME"), ""},
		[]string{path, ""},
		[]string{os.Getenv("ProgramFiles(x86)"), "/VMware/VMware Workstation"},
		[]string{os.Getenv("ProgramFiles"), "/VMware/VMware Workstation"},
	}

	return findFile(file, paths)
}

func findDataFile(file string) string {
	path, _ := workstationVmnetDhcpLeasesPathFromRegistry()

	if path != "" {
		path = filepath.Dir(path)
	}

	paths := Paths{
		[]string{os.Getenv("VMWARE_DATA"), ""},
		[]string{path, ""},
		[]string{os.Getenv("ProgramData"), "/VMWare"},
		[]string{os.Getenv("ALLUSERSPROFILE"), "/VMWare"},
	}

	return findFile(file, paths)
}

func workstationVmnetnatConfPath() string {
	const VMNETNAT_CONF = "vmnetnat.conf"

	return findDataFile(VMNETNAT_CONF)
}
