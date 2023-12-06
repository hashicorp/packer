// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"

	"github.com/hashicorp/packer/version"
)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	Meta

	CheckFunc VersionCheckFunc
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
	return "Prints the Packer version, and checks for new release."
}

func (c *VersionCommand) Run(args []string) int {
	c.Ui.Machine("version", version.Version)
	c.Ui.Machine("version-prelease", version.VersionPrerelease)
	c.Ui.Machine("version-commit", version.GitCommit)

	c.Ui.Say(fmt.Sprintf("Packer v%s", version.FormattedVersion()))

	// If we have a version check function, then let's check for
	// the latest version as well.
	if c.CheckFunc != nil {

		// Check the latest version
		info, err := c.CheckFunc()
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"\nError checking latest version: %s", err))
		}
		if info.Outdated {
			c.Ui.Say(fmt.Sprintf(
				"\nYour version of Packer is out of date! The latest version\n"+
					"is %s. You can update by downloading from www.packer.io/downloads",
				info.Latest))
		}
	}

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the Packer version"
}
