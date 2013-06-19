package validate

import (
	"flag"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"strings"
)

type Command byte

func (Command) Help() string {
	return strings.TrimSpace(helpString)
}

func (c Command) Run(env packer.Environment, args []string) int {
	var cfgSyntaxOnly bool

	cmdFlags := flag.NewFlagSet("validate", flag.ContinueOnError)
	cmdFlags.Usage = func() { env.Ui().Say(c.Help()) }
	cmdFlags.BoolVar(&cfgSyntaxOnly, "syntax-only", false, "check syntax only")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	args = cmdFlags.Args()
	if len(args) != 1 {
		cmdFlags.Usage()
		return 1
	}

	// Read the file into a byte array so that we can parse the template
	log.Printf("Reading template: %s", args[0])
	tplData, err := ioutil.ReadFile(args[0])
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Failed to read template file: %s", err))
		return 1
	}

	// Parse the template into a machine-usable format
	log.Println("Parsing template...")
	tpl, err := packer.ParseTemplate(tplData)
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	if cfgSyntaxOnly {
		env.Ui().Say("Syntax-only check passed. Everything looks okay.")
		return 0
	}

	errs := make([]error, 0)

	// The component finder for our builds
	components := &packer.ComponentFinder{
		Builder:       env.Builder,
		Hook:          env.Hook,
		PostProcessor: env.PostProcessor,
		Provisioner:   env.Provisioner,
	}

	// Otherwise, get all the builds
	buildNames := tpl.BuildNames()
	builds := make([]packer.Build, 0, len(buildNames))
	for _, buildName := range buildNames {
		log.Printf("Creating build from template for: %s", buildName)
		build, err := tpl.Build(buildName, components)
		if err != nil {
			errs = append(errs, fmt.Errorf("Build '%s': %s", buildName, err))
			continue
		}

		builds = append(builds, build)
	}

	// Check the configuration of all builds
	for _, b := range builds {
		log.Printf("Preparing build: %s", b.Name())
		err := b.Prepare()
		if err != nil {
			errs = append(errs, fmt.Errorf("Errors validating build '%s'. %s", b.Name(), err))
		}
	}

	if len(errs) > 0 {
		env.Ui().Error("Template validation failed. Errors are shown below.\n")
		for i, err := range errs {
			env.Ui().Error(err.Error())

			if (i + 1) < len(errs) {
				env.Ui().Error("")
			}
		}

		return 1
	}

	env.Ui().Say("Template validated successfully.")
	return 0
}

func (Command) Synopsis() string {
	return "check that a template is valid"
}
