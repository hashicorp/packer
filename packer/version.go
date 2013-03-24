package packer

// The version of packer.
const Version = "0.1.0.dev"

type versionCommand byte

// Implement the Command interface by simply showing the version
func (versionCommand) Run(env *Environment, args []string) int {
	env.Ui().Say("Packer v%v\n", Version)
	return 0
}
