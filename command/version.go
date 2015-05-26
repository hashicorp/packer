package command

import (
	"bytes"
	"fmt"
)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	Meta

	Revision          string
	Version           string
	VersionPrerelease string
	CheckFunc         VersionCheckFunc
}

// VersionCheckFunc is the callback called by the Version command to
// check if there is a new version of Packer.
type VersionCheckFunc func() (VersionCheckInfo, error)

// VersionCheckInfo is the return value for the VersionCheckFunc callback
// and tells the Version command information about the latest version
// of Packer.
type VersionCheckInfo struct {
	Outdated bool
	Latest   string
	Alerts   []string
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Run(args []string) int {
	c.Ui.Machine("version", c.Version)
	c.Ui.Machine("version-prelease", c.VersionPrerelease)
	c.Ui.Machine("version-commit", c.Revision)

	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Packer v%s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.Ui.Say(versionString.String())

	// If we have a version check function, then let's check for
	// the latest version as well.
	if c.CheckFunc != nil {
		// Separate the prior output with a newline
		c.Ui.Say("")

		// Check the latest version
		info, err := c.CheckFunc()
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Error checking latest version: %s", err))
		}
		if info.Outdated {
			c.Ui.Say(fmt.Sprintf(
				"Your version of Packer is out of date! The latest version\n"+
					"is %s. You can update by downloading from www.packer.io",
				info.Latest))
		}
	}

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the Packer version"
}
