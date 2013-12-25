package iso

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
)

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver(config *config) (vmwcommon.Driver, error) {
	drivers := []vmwcommon.Driver{}

	if config.RemoteType != "" {
		drivers = []vmwcommon.Driver{
			&ESX5Driver{
				Host:      config.RemoteHost,
				Port:      config.RemotePort,
				Username:  config.RemoteUser,
				Password:  config.RemotePassword,
				Datastore: config.RemoteDatastore,
			},
		}
	} else {
		switch runtime.GOOS {
		case "darwin":
			drivers = []vmwcommon.Driver{
				&Fusion5Driver{
					AppPath:   "/Applications/VMware Fusion.app",
					SSHConfig: &config.SSHConfig,
				},
			}
		case "linux":
			drivers = []vmwcommon.Driver{
				&Workstation9Driver{
					SSHConfig: &config.SSHConfig,
				},
				&Player5LinuxDriver{
					SSHConfig: &config.SSHConfig,
				},
			}
		case "windows":
			drivers = []vmwcommon.Driver{
				&Workstation9Driver{
					SSHConfig: &config.SSHConfig,
				},
			}
		default:
			return nil, fmt.Errorf("can't find driver for OS: %s", runtime.GOOS)
		}
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
