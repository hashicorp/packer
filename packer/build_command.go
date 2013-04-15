package packer

type buildCommand byte

func (buildCommand) Run(env *Environment, args []string) int {
	return 0
}

func (buildCommand) Synopsis() string {
	return "build machines images from Packer template"
}
