package virtualbox

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// A driver is able to talk to VirtualBox and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the VirtualBox builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {
	// Create a SATA controller.
	CreateSATAController(vm string, controller string) error

	// Checks if the VM with the given name is running.
	IsRunning(string) (bool, error)

	// Stop stops a running machine, forcefully.
	Stop(string) error

	// SuppressMessages should do what needs to be done in order to
	// suppress any annoying popups from VirtualBox.
	SuppressMessages() error

	// VBoxManage executes the given VBoxManage command
	VBoxManage(...string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of VirtualBox that is installed.
	Version() (string, error)
}

type VBox42Driver struct {
	// This is the path to the "VBoxManage" application.
	VBoxManagePath string
}

func (d *VBox42Driver) CreateSATAController(vmName string, name string) error {
	version, err := d.Version()
	if err != nil {
		return err
	}

	portCountArg := "--sataportcount"
	if strings.HasPrefix(version, "4.3") {
		portCountArg = "--portcount"
	}

	command := []string{
		"storagectl", vmName,
		"--name", name,
		"--add", "sata",
		portCountArg, "1",
	}

	return d.VBoxManage(command...)
}

func (d *VBox42Driver) IsRunning(name string) (bool, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.VBoxManagePath, "showvminfo", name, "--machinereadable")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return false, err
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		// Need to trim off CR character when running in windows
		line = strings.TrimRight(line, "\r")

		if line == `VMState="running"` {
			return true, nil
		}

		// We consider "stopping" to still be running. We wait for it to
		// be completely stopped or some other state.
		if line == `VMState="stopping"` {
			return true, nil
		}

		// We consider "paused" to still be running. We wait for it to
		// be completely stopped or some other state.
		if line == `VMState="paused"` {
			return true, nil
		}
	}

	return false, nil
}

func (d *VBox42Driver) Stop(name string) error {
	if err := d.VBoxManage("controlvm", name, "poweroff"); err != nil {
		return err
	}

	return nil
}

func (d *VBox42Driver) SuppressMessages() error {
	extraData := map[string]string{
		"GUI/RegistrationData": "triesLeft=0",
		"GUI/SuppressMessages": "confirmInputCapture,remindAboutAutoCapture,remindAboutMouseIntegrationOff,remindAboutMouseIntegrationOn,remindAboutWrongColorDepth",
		"GUI/UpdateDate":       fmt.Sprintf("1 d, %d-01-01, stable", time.Now().Year()+1),
		"GUI/UpdateCheckCount": "60",
	}

	for k, v := range extraData {
		if err := d.VBoxManage("setextradata", "global", k, v); err != nil {
			return err
		}
	}

	return nil
}

func (d *VBox42Driver) VBoxManage(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing VBoxManage: %#v", args)
	cmd := exec.Command(d.VBoxManagePath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("VBoxManage error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}

func (d *VBox42Driver) Verify() error {
	return nil
}

func (d *VBox42Driver) Version() (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.VBoxManagePath, "--version")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(stdout.String())
	log.Printf("VBoxManage --version output: %s", versionOutput)

	// If the "--version" output contains vboxdrv, then this is indicative
	// of problems with the VirtualBox setup and we shouldn't really continue,
	// whether or not we can read the version.
	if strings.Contains(versionOutput, "vboxdrv") {
		return "", fmt.Errorf("VirtualBox is not properly setup: %s", versionOutput)
	}

	versionRe := regexp.MustCompile("[^.0-9]")
	matches := versionRe.Split(versionOutput, 2)
	if len(matches) == 0 || matches[0] == "" {
		return "", fmt.Errorf("No version found: %s", versionOutput)
	}

	log.Printf("VirtualBox version: %s", matches[0])
	return matches[0], nil
}
