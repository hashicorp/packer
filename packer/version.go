package packer

import (
	"bytes"
	"fmt"
)

// The git commit that is being compiled. This will be filled in by the
// compiler for source builds.
var GitCommit string

// The version of packer.
const Version = "0.2.1"

// Any pre-release marker for the version. If this is "" (empty string),
// then it means that it is a final release. Otherwise, this is the
// pre-release marker.
const VersionPrerelease = "dev"

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

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	env.Ui().Say(versionString.String())
	return 0
}

func (versionCommand) Synopsis() string {
	return "print Packer version"
}
