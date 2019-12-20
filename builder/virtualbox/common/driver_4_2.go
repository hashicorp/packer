package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	versionUtil "github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/common/retry"
)

type VBox42Driver struct {
	// This is the path to the "VBoxManage" application.
	VBoxManagePath string
}

func (d *VBox42Driver) CreateSATAController(vmName string, name string, portcount int) error {
	version, err := d.Version()
	if err != nil {
		return err
	}

	portCountArg := "--portcount"

	currentVersion, err := versionUtil.NewVersion(version)
	if err != nil {
		return err
	}
	firstVersionUsingPortCount, err := versionUtil.NewVersion("4.3")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(firstVersionUsingPortCount) {
		portCountArg = "--sataportcount"
	}

	command := []string{
		"storagectl", vmName,
		"--name", name,
		"--add", "sata",
		portCountArg, strconv.Itoa(portcount),
	}

	return d.VBoxManage(command...)
}

func (d *VBox42Driver) CreateNVMeController(vmName string, name string, portcount int) error {
	command := []string{
		"storagectl", vmName,
		"--name", name,
		"--add", "pcie",
		"--controller", "NVMe",
		"--portcount", strconv.Itoa(portcount),
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
	ctx := context.TODO()
	err := retry.Config{
		Tries: 5,
		ShouldRetry: func(err error) bool {
			return strings.Contains(err.Error(), "VBOX_E_INVALID_OBJECT_STATE")
		},
		RetryDelay: func() time.Duration { return 2 * time.Minute },
	}.Run(ctx, func(ctx context.Context) error {
		_, err := d.VBoxManageWithOutput(args...)
		return err
	})

	return err
}

func (d *VBox42Driver) VBoxManageWithOutput(args ...string) (string, error) {
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

	return stdoutString, err
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

// LoadSnapshots load the snapshots for a VM instance
func (d *VBox42Driver) LoadSnapshots(vmName string) (*VBoxSnapshot, error) {
	if vmName == "" {
		panic("Argument empty exception: vmName")
	}
	log.Printf("Executing LoadSnapshots: VM: %s", vmName)

	var rootNode *VBoxSnapshot
	stdoutString, err := d.VBoxManageWithOutput("snapshot", vmName, "list", "--machinereadable")
	if stdoutString == "This machine does not have any snapshots" {
		return rootNode, nil
	}
	if nil != err {
		return nil, err
	}

	rootNode, err = ParseSnapshotData(stdoutString)
	if nil != err {
		return nil, err
	}

	return rootNode, nil
}

func (d *VBox42Driver) CreateSnapshot(vmname string, snapshotName string) error {
	if vmname == "" {
		panic("Argument empty exception: vmname")
	}
	log.Printf("Executing CreateSnapshot: VM: %s, SnapshotName %s", vmname, snapshotName)

	return d.VBoxManage("snapshot", vmname, "take", snapshotName)
}

func (d *VBox42Driver) HasSnapshots(vmname string) (bool, error) {
	if vmname == "" {
		panic("Argument empty exception: vmname")
	}
	log.Printf("Executing HasSnapshots: VM: %s", vmname)

	sn, err := d.LoadSnapshots(vmname)
	if nil != err {
		return false, err
	}
	return nil != sn, nil
}

func (d *VBox42Driver) GetCurrentSnapshot(vmname string) (*VBoxSnapshot, error) {
	if vmname == "" {
		panic("Argument empty exception: vmname")
	}
	log.Printf("Executing GetCurrentSnapshot: VM: %s", vmname)

	sn, err := d.LoadSnapshots(vmname)
	if nil != err {
		return nil, err
	}
	return sn.GetCurrentSnapshot(), nil
}

func (d *VBox42Driver) SetSnapshot(vmname string, sn *VBoxSnapshot) error {
	if vmname == "" {
		panic("Argument empty exception: vmname")
	}
	if nil == sn {
		panic("Argument null exception: sn")
	}
	log.Printf("Executing SetSnapshot: VM: %s, SnapshotName %s", vmname, sn.UUID)

	return d.VBoxManage("snapshot", vmname, "restore", sn.UUID)
}

func (d *VBox42Driver) DeleteSnapshot(vmname string, sn *VBoxSnapshot) error {
	if vmname == "" {
		panic("Argument empty exception: vmname")
	}
	if nil == sn {
		panic("Argument null exception: sn")
	}
	log.Printf("Executing DeleteSnapshot: VM: %s, SnapshotName %s", vmname, sn.UUID)
	return d.VBoxManage("snapshot", vmname, "delete", sn.UUID)
}
