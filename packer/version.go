package packer

import (
	"bytes"
	"fmt"
)

// The git commit that is being compiled. This will be filled in by the
// compiler for source builds.
var GitCommit string

// This should be check to a callback to check for the latest version.
//
// The global nature of this variable is dirty, but a version checker
// really shouldn't change anyways.
var VersionChecker VersionCheckFunc

// The version of packer.
const Version = "0.7.2"

// Any pre-release marker for the version. If this is "" (empty string),
// then it means that it is a final release. Otherwise, this is the
// pre-release marker.
const VersionPrerelease = ""

// VersionCheckFunc is the callback that is called to check the latest
// version of Packer.
type VersionCheckFunc func(string) (VersionCheckInfo, error)

// VersionCheckInfo is the return value for the VersionCheckFunc that
// contains the latest version information.
type VersionCheckInfo struct {
	Outdated bool
	Latest   string
	Alerts   []string
}

type versionCommand byte

func (versionCommand) Help() string {
	return `usage: packer version

Outputs the version of Packer that is running. There are no additional
command-line flags for this command.`
}

func (versionCommand) Run(env Environment, args []string) int {
	env.Ui().Machine("version", Version)
	env.Ui().Machine("version-prelease", VersionPrerelease)
	env.Ui().Machine("version-commit", GitCommit)
	env.Ui().Say(VersionString())

	if VersionChecker != nil {
		current := Version
		if VersionPrerelease != "" {
			current += fmt.Sprintf(".%s", VersionPrerelease)
		}

		info, err := VersionChecker(current)
		if err != nil {
			env.Ui().Say(fmt.Sprintf("\nError checking latest version: %s", err))
		}
		if info.Outdated {
			env.Ui().Say(fmt.Sprintf(
				"\nYour version of Packer is out of date! The latest version\n"+
					"is %s. You can update by downloading from www.packer.io.",
				info.Latest))
		}
	}

	return 0
}

func (versionCommand) Synopsis() string {
	return "print Packer version"
}

// VersionString returns the Packer version in human-readable
// form complete with pre-release and git commit info if it is
// available.
func VersionString() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Packer v%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", VersionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	return versionString.String()
}
