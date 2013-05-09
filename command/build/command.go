package build

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
)

type Command byte

func (Command) Run(env packer.Environment, args []string) int {
	if len(args) != 1 {
		env.Ui().Error("A single template argument is required.\n")
		return 1
	}

	// Read the file into a byte array so that we can parse the template
	log.Printf("Reading template: %s\n", args[0])
	tplData, err := ioutil.ReadFile(args[0])
	if err != nil {
		env.Ui().Error("Failed to read template file: %s\n", err.Error())
		return 1
	}

	// Parse the template into a machine-usable format
	log.Println("Parsing template...")
	tpl, err := packer.ParseTemplate(tplData)
	if err != nil {
		env.Ui().Error("Failed to parse template: %s\n", err.Error())
		return 1
	}

	// Go through each builder and compile the builds that we care about
	buildNames := tpl.BuildNames()
	builds := make([]packer.Build, 0, len(buildNames))
	for _, buildName := range buildNames {
		log.Printf("Creating build: %s\n", buildName)
		build, err := tpl.Build(buildName, env.Builder)
		if err != nil {
			env.Ui().Error("Failed to create build '%s': \n\n%s\n", buildName, err.Error())
			return 1
		}

		builds = append(builds, build)
	}

	// Prepare all the builds
	for _, b := range builds {
		log.Printf("Preparing build: %s\n", b.Name())
		err := b.Prepare()
		if err != nil {
			env.Ui().Error("%s\n", err)
			return 1
		}
	}

	env.Ui().Say("YAY!\n")
	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from template"
}
