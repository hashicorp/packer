package packer

import (
	"io/ioutil"
)

type buildCommand byte

func (buildCommand) Run(env Environment, args []string) int {
	if len(args) != 1 {
		// TODO: Error message
		return 1
	}

	// Read the file into a byte array so that we can parse the template
	tplData, err := ioutil.ReadFile(args[0])
	if err != nil {
		// TODO: Error message
		return 1
	}

	// Parse the template into a machine-usable format
	_, err = ParseTemplate(tplData)
	if err != nil {
		// TODO: error message
		return 1
	}

	// Go through each builder and compile the builds that we care about
	//builds := make([]Build, 0, len(tpl.Builders))
	//for name, rawConfig := range tpl.Builders {
		//builder := env.Builder(name, rawConfig)
		//build := env.Build(name, builder)
		//builds = append(builds, build)
	//}

	return 0
}

func (buildCommand) Synopsis() string {
	return "build machines images from Packer template"
}
