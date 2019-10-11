package common

// Parallels10Driver are inherited from Parallels9Driver.
type Parallels10Driver struct {
	Parallels9Driver
}

// SetDefaultConfiguration applies pre-defined default settings to the VM config.
func (d *Parallels10Driver) SetDefaultConfiguration(vmName string) error {
	commands := make([][]string, 10)
	commands[0] = []string{"set", vmName, "--startup-view", "same"}
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
