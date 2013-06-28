package packer

import (
	"bytes"
	"fmt"
)

// The version of packer.
const Version = "0.1.0"

// Any pre-release marker for the version. If this is "" (empty string),
// then it means that it is a final release. Otherwise, this is the
// pre-release marker.
const VersionPrerelease = ""

type versionCommand byte

func (versionCommand) Help() string {
	return `usage: packer version

Outputs the version of Packer that is running. There are no additional
command-line flags for this command.`
}

func (versionCommand) Run(env Environment, args []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Packer v%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", VersionPrerelease)
	}

	env.Ui().Say(versionString.String())
	return 0
}

func (versionCommand) Synopsis() string {
	return "print Packer version"
}
