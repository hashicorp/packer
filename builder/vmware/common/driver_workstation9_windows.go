// +build windows

package common

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

	return findFile("vmware-vdiskmanager.exe", workstationProgramFilePaths()), nil
}

func workstationFindVMware() (string, error) {
	path, err := exec.LookPath("vmware.exe")
	if err == nil {
		return path, nil
	}

	return findFile("vmware.exe", workstationProgramFilePaths()), nil
}

func workstationFindVmrun() (string, error) {
	path, err := exec.LookPath("vmrun.exe")
	if err == nil {
		return path, nil
	}

	return findFile("vmrun.exe", workstationProgramFilePaths()), nil
}

func workstationToolsIsoPath(flavor string) string {
	return findFile(flavor+".iso", workstationProgramFilePaths())
}

func workstationDhcpLeasesPath(device string) string {
	path, err := workstationDhcpLeasesPathRegistry()
	if err != nil {
		log.Printf("Error finding leases in registry: %s", err)
	} else if _, err := os.Stat(path); err == nil {
		return path
	}

	return findFile("vmnetdhcp.leases", workstationDataFilePaths())
}

func workstationVmnetnatConfPath() string {
	return findFile("vmnetnat.conf", workstationDataFilePaths())
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
func workstationDhcpLeasesPathRegistry() (s string, err error) {
	key := "SYSTEM\\CurrentControlSet\\services\\VMnetDHCP\\Parameters"
	subkey := "LeaseFile"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

func normalizePath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.TrimRight(path, "/")
	return path
}

func findFile(file string, paths []string) string {
	for _, path := range paths {
		path = filepath.Join(path, file)
		path = normalizePath(path)
		log.Printf("Searching for file '%s'", path)

		if _, err := os.Stat(path); err == nil {
			log.Printf("Found file '%s'", path)
			return path
		}
	}

	log.Printf("File not found: '%s'", file)
	return ""
}

// workstationProgramFilesPaths returns a list of paths that are eligible
// to contain program files we may want just as vmware.exe.
func workstationProgramFilePaths() []string {
	path, err := workstationVMwareRoot()
	if err != nil {
		log.Printf("Error finding VMware root: %s", err)
	}

	paths := make([]string, 0, 5)
	if os.Getenv("VMWARE_HOME") != "" {
		paths = append(paths, os.Getenv("VMWARE_HOME"))
	}

	if path != "" {
		paths = append(paths, path)
	}

	if os.Getenv("ProgramFiles(x86)") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "/VMware/VMware Workstation"))
	}

	if os.Getenv("ProgramFiles") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "/VMware/VMware Workstation"))
	}

	return paths
}

// workstationDataFilePaths returns a list of paths that are eligible
// to contain data files we may want such as vmnet NAT configuration files.
func workstationDataFilePaths() []string {
	leasesPath, err := workstationDhcpLeasesPathRegistry()
	if err != nil {
		log.Printf("Error getting DHCP leases path: %s", err)
	}

	if leasesPath != "" {
		leasesPath = filepath.Dir(leasesPath)
	}

	paths := make([]string, 0, 5)
	if os.Getenv("VMWARE_DATA") != "" {
		paths = append(paths, os.Getenv("VMWARE_DATA"))
	}

	if leasesPath != "" {
		paths = append(paths, leasesPath)
	}

	if os.Getenv("ProgramData") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramData"), "/VMware"))
	}

	if os.Getenv("ALLUSERSPROFILE") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ALLUSERSPROFILE"), "/Application Data/VMware"))
	}

	return paths
}
