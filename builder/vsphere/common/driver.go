package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// Clones the VM to another VM
	CloneVirtualMachine(string, string, string, string, uint, uint, uint, bool, string, string, string) error

	// Create a new VM
	CreateVirtualMachine(string, string, string, uint, uint, uint, bool, string, string, string, string) error

	// Destroys a VM
	Destroy() error

	//Create and attach an additional disk
	CreateDisk(uint, bool) error

	// Attach the VMware tools ISO
	ToolsInstall() error

	// Checks if the VM is started
	IsRunning() (bool, error)

	// Start starts a VM specified by the path to the VMX given.
	Start() error

	// Checks if the VM is stopped.
	IsStopped() (bool, error)

	// Stop stops a VM
	Stop() error

	// Checks if the VM is destroyed.
	IsDestroyed() (bool, error)

	// Upload uploads a local file (iso, floppy img) to the remote side and returns the
	// new path that should be used in the driver along with an error if it exists.
	Upload(string, string) (string, error)

	// Export the VM from the remote side to local path
	ExportVirtualMachine(string, string, []string) error

	// Add floppy device and insert the image in the provided path
	// return the created device name
	AddFloppy(string) (string, error)

	// Remove the floppy drive
	RemoveFloppy(floppyDevice string) error

	// Add cdrom device and insert the iso in the provided path
	// return the created device name
	MountISO(string) (string, error)

	// Remove the cdrom
	UnmountISO(string) error

	// Change parameter (provided in a string key=value) of the VM
	VMChange(string) error

	// Disable VNC on the VM
	VNCDisable() error

	// Enable VNC on the VM return the host and port for VNC connection
	VNCEnable(string, uint, uint) (string, uint, error)

	// VM Ip
	GuestIP() (string, error)

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error.
	Verify() error
}

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver(dconfig *DriverConfig, config *SSHConfig) (Driver, error) {

	drivers := []Driver{
		&GOVCDriver{
			Vcenter:        dconfig.Vcenter,
			Host:           dconfig.RemoteHost,
			Datacenter:     dconfig.RemoteDatacenter,
			Cluster:        dconfig.RemoteCluster,
			ResourcePool:   dconfig.RemoteResourcePool,
			Username:       dconfig.RemoteUser,
			Password:       dconfig.RemotePassword,
			Insecure:       dconfig.Insecure,
			Datastore:      dconfig.RemoteDatastore,
			CacheDatastore: dconfig.RemoteCacheDatastore,
			CacheFolder:    dconfig.RemoteCacheDirectory,
			SSHConfig:      config,
		},
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

		err = fmt.Errorf("VSphere error: %s", message)

	}

	//Reduce log size to avoid the never ending JSON
	if len(stdoutString) > 200 {
		stdoutString = fmt.Sprintf("%s ...", stdoutString[0:200])
	}
	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	// Replace these for Windows, we only want to deal with Unix
	// style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}
