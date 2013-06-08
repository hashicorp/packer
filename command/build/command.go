package build

import (
	"flag"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type Command byte

func (Command) Help() string {
	return strings.TrimSpace(helpText)
}

func (c Command) Run(env packer.Environment, args []string) int {
	var cfgOnly []string

	cmdFlags := flag.NewFlagSet("build", flag.ContinueOnError)
	cmdFlags.Usage = func() { env.Ui().Say(c.Help()) }
	cmdFlags.Var((*stringSliceValue)(&cfgOnly), "only", "only build the given builds by name")
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
		if len(cfgOnly) > 0 {
			found := false
			for _, only := range cfgOnly {
				if buildName == only {
					found = true
					break
				}
			}

			if !found {
				log.Printf("Skipping build '%s' because not specified by -only.", buildName)
				continue
			}
		}

		log.Printf("Creating build: %s", buildName)
		build, err := tpl.Build(buildName, components)
		if err != nil {
			env.Ui().Error(fmt.Sprintf("Failed to create build '%s': \n\n%s", buildName, err))
			return 1
		}

		builds = append(builds, build)
	}

	// Compile all the UIs for the builds
	colors := [5]packer.UiColor{
		packer.UiColorGreen,
		packer.UiColorYellow,
		packer.UiColorBlue,
		packer.UiColorMagenta,
		packer.UiColorCyan,
	}

	buildUis := make(map[string]packer.Ui)
	for i, b := range builds {
		var ui packer.Ui

		ui = &packer.ColoredUi{
			colors[i%len(colors)],
			env.Ui(),
		}

		ui = &packer.PrefixedUi{
			fmt.Sprintf("==> %s", b.Name()),
			fmt.Sprintf("    %s", b.Name()),
			ui,
		}

		buildUis[b.Name()] = ui
		ui.Say(fmt.Sprintf("%s output will be in this color.", b.Name()))
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
		// Increment the waitgroup so we wait for this item to finish properly
		wg.Add(1)

		// Run the build in a goroutine
		go func(b packer.Build) {
			defer wg.Done()

			log.Printf("Starting build run: %s", b.Name())
			ui := buildUis[b.Name()]
			artifacts[b.Name()] = b.Run(ui)
			ui.Say("Build finished.")
		}(b)
	}

	// Handle signals
	var interruptWg sync.WaitGroup
	interrupted := false
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		<-sigCh
		interruptWg.Add(1)
		defer interruptWg.Done()
		interrupted = true

		log.Println("Interrupted! Cancelling builds...")

		var wg sync.WaitGroup
		for _, b := range builds {
			wg.Add(1)

			go func(b packer.Build) {
				defer wg.Done()

				log.Printf("Stopping build: %s", b.Name())
				b.Cancel()
				log.Printf("Build cancelled: %s", b.Name())
			}(b)
		}

		wg.Wait()
	}()

	// Wait for both the builds to complete and the interrupt handler,
	// if it is interrupted.
	log.Printf("Waiting on builds to complete...")
	wg.Wait()

	log.Printf("Builds completed. Waiting on interrupt barrier...")
	interruptWg.Wait()
	log.Printf("Interrupt barrier passed.")

	if interrupted {
		env.Ui().Say("Cleanly cancelled builds after being interrupted.")
		return 1
	}

	// Output all the artifacts
	env.Ui().Say("\n==> The build completed! The artifacts created were:")
	for name, artifact := range artifacts {
		env.Ui().Say(fmt.Sprintf("--> %s:", name))

		if artifact != nil {
			env.Ui().Say(artifact.String())
		} else {
			env.Ui().Say("<nothing>")
		}
	}

	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from template"
}
