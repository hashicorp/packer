package build

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"sync"
)

type Command byte

func (Command) Run(env packer.Environment, args []string) int {
	if len(args) != 1 {
		env.Ui().Error("A single template argument is required.")
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

	// The component finder for our builds
	components := &packer.ComponentFinder{
		Builder:     env.Builder,
		Hook:        env.Hook,
		Provisioner: env.Provisioner,
	}

	// Go through each builder and compile the builds that we care about
	buildNames := tpl.BuildNames()
	builds := make([]packer.Build, 0, len(buildNames))
	for _, buildName := range buildNames {
		log.Printf("Creating build: %s", buildName)
		build, err := tpl.Build(buildName, components)
		if err != nil {
			env.Ui().Error(fmt.Sprintf("Failed to create build '%s': \n\n%s", buildName, err))
			return 1
		}

		builds = append(builds, build)
	}

	// Compile all the UIs for the builds
	buildUis := make(map[string]packer.Ui)
	for _, b := range builds {
		buildUis[b.Name()] = &packer.PrefixedUi{
			fmt.Sprintf("==> %s", b.Name()),
			env.Ui(),
		}
	}

	// Prepare all the builds
	for _, b := range builds {
		log.Printf("Preparing build: %s", b.Name())
		err := b.Prepare(buildUis[b.Name()])
		if err != nil {
			env.Ui().Error(err.Error())
			return 1
		}
	}

	// Run all the builds in parallel and wait for them to complete
	var wg sync.WaitGroup
	artifacts := make(map[string]packer.Artifact)
	for _, b := range builds {
		log.Printf("Starting build run: %s", b.Name())

		// Increment the waitgroup so we wait for this item to finish properly
		wg.Add(1)

		// Run the build in a goroutine
		go func() {
			defer wg.Done()
			artifacts[b.Name()] = b.Run(buildUis[b.Name()])
		}()
	}

	wg.Wait()

	// Output all the artifacts
	env.Ui().Say("\n==> The build completed! The artifacts created were:")
	for name, artifact := range artifacts {
		env.Ui().Say(fmt.Sprintf("--> %s:", name))
		env.Ui().Say(artifact.String())
	}

	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from template"
}
