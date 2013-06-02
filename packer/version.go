package packer

import "fmt"

// The version of packer.
const Version = "0.1.0.dev"

type versionCommand byte

func (versionCommand) Help() string {
	return `usage: packer version

Outputs the version of Packer that is running. There are no additional
command-line flags for this command.`
}

func (versionCommand) Run(env Environment, args []string) int {
	env.Ui().Say(fmt.Sprintf("Packer v%v\n", Version))
	return 0
}

func (versionCommand) Synopsis() string {
	return "print Packer version"
}
