package vmware

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CompactDisk compacts a virtual disk.
	CompactDisk(string) error

	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string, bool) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// Get the path to the VMware ISO for the given flavor.
	ToolsIsoPath(string) string

	// Get the path to the DHCP leases file for the given device.
	DhcpLeasesPath(string) string

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error.
	Verify() error
}

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver() (Driver, error) {
	drivers := []Driver{}

	switch runtime.GOOS {
	case "darwin":
		drivers = []Driver{
			&Fusion5Driver{
				AppPath: "/Applications/VMware Fusion.app",
			},
		}
	case "linux":
		drivers = []Driver{
			new(Workstation9Driver),
			new(Player5LinuxDriver),
		}
	case "windows":
		drivers = []Driver{
			new(Workstation9Driver),
		}
	default:
		return nil, fmt.Errorf("can't find driver for OS: %s", runtime.GOOS)
	}

	errs := ""
	for _, driver := range drivers {
		err := driver.Verify()
		if err == nil {
			return driver, nil
		}
		errs += "* " + err.Error() + "\n"
	}

	return nil, fmt.Errorf(
		"Unable to initialize any driver for this platform. The errors\n"+
			"from each driver are shown below. Please fix at least one driver\n"+
			"to continue:\n%s", errs)
}

func runAndLog(cmd *exec.Cmd) (string, string, error) {
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

	// Replace these for Windows, we only want to deal with Unix
	// style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}
