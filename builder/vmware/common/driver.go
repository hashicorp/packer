package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/mitchellh/multistep"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// Clone clones the VMX and the disk to the destination path. The
	// destination is a path to the VMX file. The disk will be copied
	// to that same directory.
	Clone(dst string, src string) error

	// CompactDisk compacts a virtual disk.
	CompactDisk(string) error

	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// CommHost returns the host address for the VM that is being
	// managed by this driver.
	CommHost(multistep.StateBag) (string, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string, bool) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// SuppressMessages modifies the VMX or surrounding directory so that
	// VMware doesn't show any annoying messages.
	SuppressMessages(string) error

	// Get the path to the VMware ISO for the given flavor.
	ToolsIsoPath(string) string

	// Attach the VMware tools ISO
	ToolsInstall() error

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
func NewDriver(dconfig *DriverConfig, config *SSHConfig) (Driver, error) {
	drivers := []Driver{}

	switch runtime.GOOS {
	case "darwin":
		drivers = []Driver{
			&Fusion6Driver{
				Fusion5Driver: Fusion5Driver{
					AppPath:   dconfig.FusionAppPath,
					SSHConfig: config,
				},
			},
			&Fusion5Driver{
				AppPath:   dconfig.FusionAppPath,
				SSHConfig: config,
			},
		}
	case "linux":
		fallthrough
	case "windows":
		drivers = []Driver{
			&Workstation10Driver{
				Workstation9Driver: Workstation9Driver{
					SSHConfig: config,
				},
			},
			&Workstation9Driver{
				SSHConfig: config,
			},
			&Player6Driver{
				Player5Driver: Player5Driver{
					SSHConfig: config,
				},
			},
			&Player5Driver{
				SSHConfig: config,
			},
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
		message := stderrString
		if message == "" {
			message = stdoutString
		}

		err = fmt.Errorf("VMware error: %s", message)

		// If "unknown error" is in there, add some additional notes
		re := regexp.MustCompile(`(?i)unknown error`)
		if re.MatchString(message) {
			err = fmt.Errorf(
				"%s\n\n%s", err,
				"Packer detected a VMware 'Unknown Error'. Unfortunately VMware\n"+
					"often has extremely vague error messages such as this and Packer\n"+
					"itself can't do much about that. Please check the vmware.log files\n"+
					"created by VMware when a VM is started (in the directory of the\n"+
					"vmx file), which often contains more detailed error information.")
		}
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	// Replace these for Windows, we only want to deal with Unix
	// style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}

func normalizeVersion(version string) (string, error) {
	i, err := strconv.Atoi(version)
	if err != nil {
		return "", fmt.Errorf(
			"VMware version '%s' is not numeric", version)
	}

	return fmt.Sprintf("%02d", i), nil
}

func compareVersions(versionFound string, versionWanted string, product string) error {
	found, err := normalizeVersion(versionFound)
	if err != nil {
		return err
	}

	wanted, err := normalizeVersion(versionWanted)
	if err != nil {
		return err
	}

	if found < wanted {
		return fmt.Errorf(
			"VMware %s version %s, or greater, is required. Found version: %s", product, versionWanted, versionFound)
	}

	return nil
}
