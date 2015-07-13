package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

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
	if strings.HasPrefix(version, "4.3") || strings.HasPrefix(version, "5.") {
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

func (d *VBox42Driver) CreateSCSIController(vmName string, name string) error {

	command := []string{
		"storagectl", vmName,
		"--name", name,
		"--add", "scsi",
		"--controller", "LSILogic",
	}

	return d.VBoxManage(command...)
}

func (d *VBox42Driver) Delete(name string) error {
	return d.VBoxManage("unregistervm", name, "--delete")
}

func (d *VBox42Driver) Iso() (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.VBoxManagePath, "list", "systemproperties")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	DefaultGuestAdditionsRe := regexp.MustCompile("Default Guest Additions ISO:(.+)")

	for _, line := range strings.Split(stdout.String(), "\n") {
		// Need to trim off CR character when running in windows
		// Trimming whitespaces at this point helps to filter out empty value
		line = strings.TrimRight(line, " \r")

		matches := DefaultGuestAdditionsRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		isoname := strings.Trim(matches[1], " \r\n")
		log.Printf("Found Default Guest Additions ISO: %s", isoname)

		return isoname, nil
	}

	return "", fmt.Errorf("Cannot find \"Default Guest Additions ISO\" in vboxmanage output (or it is empty)")
}

func (d *VBox42Driver) Import(name string, path string, flags []string) error {
	args := []string{
		"import", path,
		"--vsys", "0",
		"--vmname", name,
	}
	args = append(args, flags...)

	return d.VBoxManage(args...)
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

	// We sleep here for a little bit to let the session "unlock"
	time.Sleep(2 * time.Second)

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

	if err == nil {
		// Sometimes VBoxManage gives us an error with a zero exit code,
		// so we also regexp match an error string.
		m, _ := regexp.MatchString("VBoxManage([.a-z]+?): error:", stderrString)
		if m {
			err = fmt.Errorf("VBoxManage error: %s", stderrString)
		}
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

	versionRe := regexp.MustCompile("^([.0-9]+)(?:_(?:RC|OSEr)[0-9]+)?")
	matches := versionRe.FindAllStringSubmatch(versionOutput, 1)
	if matches == nil || len(matches[0]) != 2 {
		return "", fmt.Errorf("No version found: %s", versionOutput)
	}

	log.Printf("VirtualBox version: %s", matches[0][1])
	return matches[0][1], nil
}
