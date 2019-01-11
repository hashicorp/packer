package vagrant

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// A driver is able to talk to Vagrant and perform certain
// operations with it.

type VagrantDriver interface {
	// Calls "vagrant init"
	Init([]string) error

	// Calls "vagrant add"
	Add([]string) error

	// Calls "vagrant up"
	Up([]string) (string, string, error)

	// Calls "vagrant halt"
	Halt() error

	// Calls "vagrant suspend"
	Suspend() error

	SSHConfig() (*VagrantSSHConfig, error)

	// Calls "vagrant destroy"
	Destroy() error

	// Calls "vagrant package"[
	Package([]string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of VirtualBox that is installed.
	Version() (string, error)
}

func NewDriver() (VagrantDriver, error) {
	// Hardcode path for now while I'm developing. Obviously this path needs
	// to be discovered based on OS.
	vagrantBinary := "vagrant"
	if runtime.GOOS == "windows" {
		vagrantBinary = "vagrant.exe"
	}

	if _, err := exec.LookPath(vagrantBinary); err != nil {
		return nil, fmt.Errorf("Error: Packer cannot find Vagrant in the path: %s", err.Error())
	}

	driver := &Vagrant_2_2_Driver{
		vagrantBinary: vagrantBinary,
	}

	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}

func findVBoxManageWindows(paths string) string {
	for _, path := range strings.Split(paths, ";") {
		path = filepath.Join(path, "VBoxManage.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
