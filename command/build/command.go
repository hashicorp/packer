package build

import (
	"bytes"
	"flag"
	"fmt"
	cmdcommon "github.com/mitchellh/packer/common/command"
	"github.com/mitchellh/packer/packer"
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
	var cfgDebug bool
	var cfgForce bool
	buildOptions := new(cmdcommon.BuildOptions)

	cmdFlags := flag.NewFlagSet("build", flag.ContinueOnError)
	cmdFlags.Usage = func() { env.Ui().Say(c.Help()) }
	cmdFlags.BoolVar(&cfgDebug, "debug", false, "debug mode for builds")
	cmdFlags.BoolVar(&cfgForce, "force", false, "force a build if artifacts exist")
	cmdcommon.BuildOptionFlags(cmdFlags, buildOptions)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	args = cmdFlags.Args()
	if len(args) != 1 {
		cmdFlags.Usage()
		return 1
	}

	if err := buildOptions.Validate(); err != nil {
		env.Ui().Error(err.Error())
		env.Ui().Error("")
		env.Ui().Error(c.Help())
		return 1
	}

	// Read the file into a byte array so that we can parse the template
	log.Printf("Reading template: %s", args[0])
	tpl, err := packer.ParseTemplateFile(args[0])
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// The component finder for our builds
	components := &packer.ComponentFinder{
		Builder:       env.Builder,
		Hook:          env.Hook,
		PostProcessor: env.PostProcessor,
		Provisioner:   env.Provisioner,
	}

	// Go through each builder and compile the builds that we care about
	builds, err := buildOptions.Builds(tpl, components)
	if err != nil {
		env.Ui().Error(err.Error())
		return 1
	}

	if cfgDebug {
		env.Ui().Say("Debug mode enabled. Builds will not be parallelized.")
	}

	// Compile all the UIs for the builds
	colors := [5]packer.UiColor{
		packer.UiColorGreen,
		packer.UiColorCyan,
		packer.UiColorMagenta,
		packer.UiColorYellow,
		packer.UiColorBlue,
	}

	buildUis := make(map[string]packer.Ui)
	for i, b := range builds {
		ui := &packer.ColoredUi{
			Color: colors[i%len(colors)],
			Ui:    env.Ui(),
		}

		buildUis[b.Name()] = ui
		ui.Say(fmt.Sprintf("%s output will be in this color.", b.Name()))
	}

	// Add a newline between the color output and the actual output
	env.Ui().Say("")

	log.Printf("Build debug mode: %v", cfgDebug)
	log.Printf("Force build: %v", cfgForce)

	// Set the debug and force mode and prepare all the builds
	for _, b := range builds {
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(cfgDebug)
		b.SetForce(cfgForce)
		err := b.Prepare(nil)
		if err != nil {
			env.Ui().Error(err.Error())
			return 1
		}
	}

	// Run all the builds in parallel and wait for them to complete
	var interruptWg, wg sync.WaitGroup
	interrupted := false
	artifacts := make(map[string][]packer.Artifact)
	errors := make(map[string]error)
	for _, b := range builds {
		// Increment the waitgroup so we wait for this item to finish properly
		wg.Add(1)

		// Handle interrupts for this build
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)
		go func(b packer.Build) {
			<-sigCh
			interruptWg.Add(1)
			defer interruptWg.Done()
			interrupted = true

			log.Printf("Stopping build: %s", b.Name())
			b.Cancel()
			log.Printf("Build cancelled: %s", b.Name())
		}(b)

		// Run the build in a goroutine
		go func(b packer.Build) {
			defer wg.Done()

			name := b.Name()
			log.Printf("Starting build run: %s", name)
			ui := buildUis[name]
			runArtifacts, err := b.Run(ui, env.Cache())

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored: %s", name, err))
				errors[name] = err
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished.", name))
				artifacts[name] = runArtifacts
			}
		}(b)

		if cfgDebug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if interrupted {
			log.Println("Interrupted, not going to start any more builds.")
			break
		}
	}

	// Wait for both the builds to complete and the interrupt handler,
	// if it is interrupted.
	log.Printf("Waiting on builds to complete...")
	wg.Wait()

	log.Printf("Builds completed. Waiting on interrupt barrier...")
	interruptWg.Wait()

	if interrupted {
		env.Ui().Say("Cleanly cancelled builds after being interrupted.")
		return 1
	}

	if len(errors) > 0 {
		env.Ui().Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errors {
			env.Ui().Error(fmt.Sprintf("--> %s: %s", name, err))
		}
	}

	if len(artifacts) > 0 {
		env.Ui().Say("\n==> Builds finished. The artifacts of successful builds are:")
		for name, buildArtifacts := range artifacts {
			for _, artifact := range buildArtifacts {
				var message bytes.Buffer
				fmt.Fprintf(&message, "--> %s: ", name)

				if artifact != nil {
					fmt.Fprintf(&message, artifact.String())
				} else {
					fmt.Fprint(&message, "<nothing>")
				}

				env.Ui().Say(message.String())
			}
		}
	} else {
		env.Ui().Say("\n==> Builds finished but no artifacts were created.")
	}

	if len(errors) > 0 {
		// If any errors occurred, exit with a non-zero exit status
		return 1
	}

	return 0
}

func (Command) Synopsis() string {
	return "build image(s) from template"
}
