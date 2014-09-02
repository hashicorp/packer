// +build windows

package common

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func playerFindVdiskManager() (string, error) {
	path, err := exec.LookPath("vmware-vdiskmanager.exe")
	if err == nil {
		return path, nil
	}

	return findFile("vmware-vdiskmanager.exe", playerProgramFilePaths()), nil
}

func playerFindQemuImg() (string, error) {
	path, err := exec.LookPath("qemu-img.exe")
	if err == nil {
		return path, nil
	}

	return findFile("qemu-img.exe", playerProgramFilePaths()), nil
}

func playerFindVMware() (string, error) {
	path, err := exec.LookPath("vmplayer.exe")
	if err == nil {
		return path, nil
	}

	return findFile("vmplayer.exe", playerProgramFilePaths()), nil
}

func playerFindVmrun() (string, error) {
	path, err := exec.LookPath("vmrun.exe")
	if err == nil {
		return path, nil
	}

	return findFile("vmrun.exe", playerProgramFilePaths()), nil
}

func playerToolsIsoPath(flavor string) string {
	return findFile(flavor+".iso", playerProgramFilePaths())
}

func playerDhcpLeasesPath(device string) string {
	path, err := playerDhcpLeasesPathRegistry()
	if err != nil {
		log.Printf("Error finding leases in registry: %s", err)
	} else if _, err := os.Stat(path); err == nil {
		return path
	}

	return findFile("vmnetdhcp.leases", playerDataFilePaths())
}

func playerVmnetnatConfPath() string {
	return findFile("vmnetnat.conf", playerDataFilePaths())
}

// This reads the VMware installation path from the Windows registry.
func playerVMwareRoot() (s string, err error) {
	key := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\vmplayer.exe`
	subkey := "Path"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

// This reads the VMware DHCP leases path from the Windows registry.
func playerDhcpLeasesPathRegistry() (s string, err error) {
	key := "SYSTEM\\CurrentControlSet\\services\\VMnetDHCP\\Parameters"
	subkey := "LeaseFile"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Printf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

// playerProgramFilesPaths returns a list of paths that are eligible
// to contain program files we may want just as vmware.exe.
func playerProgramFilePaths() []string {
	path, err := playerVMwareRoot()
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
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "/VMware/VMware Player"))
	}

	if os.Getenv("ProgramFiles") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "/VMware/VMware Player"))
	}

	if os.Getenv("QEMU_HOME") != "" {
		paths = append(paths, os.Getenv("QEMU_HOME"))
	}

	if os.Getenv("ProgramFiles(x86)") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "/QEMU"))
	}

	if os.Getenv("ProgramFiles") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "/QEMU"))
	}

	if os.Getenv("SystemDrive") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("SystemDrive"), "/QEMU"))
	}

	return paths
}

// playerDataFilePaths returns a list of paths that are eligible
// to contain data files we may want such as vmnet NAT configuration files.
func playerDataFilePaths() []string {
	leasesPath, err := playerDhcpLeasesPathRegistry()
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
