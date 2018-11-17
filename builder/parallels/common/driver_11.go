package common

import (
	"fmt"
	"os/exec"
	"regexp"
)

// Parallels11Driver are inherited from Parallels9Driver.
// Used for Parallels Desktop 11, requires Pro or Business Edition
type Parallels11Driver struct {
	Parallels9Driver
}

// Verify raises an error if the builder could not be used on that host machine.
func (d *Parallels11Driver) Verify() error {

	stdout, err := exec.Command(d.PrlsrvctlPath, "info", "--license").Output()
	if err != nil {
		return err
	}

	editionRe := regexp.MustCompile(`edition="(\w+)"`)
	matches := editionRe.FindStringSubmatch(string(stdout))
	if matches == nil {
		return fmt.Errorf(
			"Could not determine your Parallels Desktop edition using: %s info --license", d.PrlsrvctlPath)
	}
	switch matches[1] {
	case "pro", "business":
		break
	default:
		return fmt.Errorf("Packer can be used only with Parallels Desktop 11 Pro or Business edition. You use: %s edition", matches[1])
	}

	return nil
}

// SetDefaultConfiguration applies pre-defined default settings to the VM config.
func (d *Parallels11Driver) SetDefaultConfiguration(vmName string) error {
	commands := make([][]string, 10)
	commands[0] = []string{"set", vmName, "--startup-view", "headless"}
	commands[1] = []string{"set", vmName, "--on-shutdown", "close"}
	commands[2] = []string{"set", vmName, "--on-window-close", "keep-running"}
	commands[3] = []string{"set", vmName, "--auto-share-camera", "off"}
	commands[4] = []string{"set", vmName, "--smart-guard", "off"}
	commands[5] = []string{"set", vmName, "--shared-cloud", "off"}
	commands[6] = []string{"set", vmName, "--shared-profile", "off"}
	commands[7] = []string{"set", vmName, "--smart-mount", "off"}
	commands[8] = []string{"set", vmName, "--sh-app-guest-to-host", "off"}
	commands[9] = []string{"set", vmName, "--sh-app-host-to-guest", "off"}

	for _, command := range commands {
		err := d.Prlctl(command...)
		if err != nil {
			return err
		}
	}
	return nil
}
