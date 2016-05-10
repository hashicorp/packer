package command

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template"
)

type BuildCommand struct {
	Meta
}

func (c BuildCommand) Run(args []string) int {
	var cfgColor, cfgDebug, cfgForce, cfgParallel bool
	flags := c.Meta.FlagSet("build", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	flags.BoolVar(&cfgColor, "color", true, "")
	flags.BoolVar(&cfgDebug, "debug", false, "")
	flags.BoolVar(&cfgForce, "force", false, "")
	flags.BoolVar(&cfgParallel, "parallel", true, "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Parse the template
	var tpl *template.Template
	var err error
	tpl, err = template.ParseFile(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// Get the core
	core, err := c.Meta.Core(tpl)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// Get the builds we care about
	buildNames := c.Meta.BuildNames(core)
	builds := make([]packer.Build, 0, len(buildNames))
	for _, n := range buildNames {
		b, err := core.Build(n)
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to initialize build '%s': %s",
				n, err))
			continue
		}

		builds = append(builds, b)
	}

	if cfgDebug {
		c.Ui.Say("Debug mode enabled. Builds will not be parallelized.")
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
	for i, b := range buildNames {
		var ui packer.Ui
		ui = c.Ui
		if cfgColor {
			ui = &packer.ColoredUi{
				Color: colors[i%len(colors)],
				Ui:    ui,
			}
		}

		buildUis[b] = ui
		ui.Say(fmt.Sprintf("%s output will be in this color.", b))
	}

	// Add a newline between the color output and the actual output
	c.Ui.Say("")

	log.Printf("Build debug mode: %v", cfgDebug)
	log.Printf("Force build: %v", cfgForce)

	// Set the debug and force mode and prepare all the builds
	for _, b := range builds {
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(cfgDebug)
		b.SetForce(cfgForce)

		warnings, err := b.Prepare()
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		if len(warnings) > 0 {
			ui := buildUis[b.Name()]
			ui.Say(fmt.Sprintf("Warnings for build '%s':\n", b.Name()))
			for _, warning := range warnings {
				ui.Say(fmt.Sprintf("* %s", warning))
			}
			ui.Say("")
		}
	}

	// Run all the builds in parallel and wait for them to complete
	var interruptWg, wg sync.WaitGroup
	interrupted := false
	var artifacts = struct {
		sync.RWMutex
		m map[string][]packer.Artifact
	}{m: make(map[string][]packer.Artifact)}
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
			runArtifacts, err := b.Run(ui, c.Cache)

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored: %s", name, err))
				errors[name] = err
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished.", name))
				artifacts.Lock()
				artifacts.m[name] = runArtifacts
				artifacts.Unlock()
			}
		}(b)

		if cfgDebug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if !cfgParallel {
			log.Printf("Parallelization disabled, waiting for build to finish: %s", b.Name())
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
		c.Ui.Say("Cleanly cancelled builds after being interrupted.")
		return 1
	}

	if len(errors) > 0 {
		c.Ui.Machine("error-count", strconv.FormatInt(int64(len(errors)), 10))

		c.Ui.Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errors {
			// Create a UI for the machine readable stuff to be targetted
			ui := &packer.TargettedUi{
				Target: name,
				Ui:     c.Ui,
			}

			ui.Machine("error", err.Error())

			c.Ui.Error(fmt.Sprintf("--> %s: %s", name, err))
		}
	}

	if len(artifacts.m) > 0 {
		c.Ui.Say("\n==> Builds finished. The artifacts of successful builds are:")
		for name, buildArtifacts := range artifacts.m {
			// Create a UI for the machine readable stuff to be targetted
			ui := &packer.TargettedUi{
				Target: name,
				Ui:     c.Ui,
			}

			// Machine-readable helpful
			ui.Machine("artifact-count", strconv.FormatInt(int64(len(buildArtifacts)), 10))

			for i, artifact := range buildArtifacts {
				var message bytes.Buffer
				fmt.Fprintf(&message, "--> %s: ", name)

				if artifact != nil {
					fmt.Fprint(&message, artifact.String())
				} else {
					fmt.Fprint(&message, "<nothing>")
				}

				iStr := strconv.FormatInt(int64(i), 10)
				if artifact != nil {
					ui.Machine("artifact", iStr, "builder-id", artifact.BuilderId())
					ui.Machine("artifact", iStr, "id", artifact.Id())
					ui.Machine("artifact", iStr, "string", artifact.String())

					files := artifact.Files()
					ui.Machine("artifact",
						iStr,
						"files-count", strconv.FormatInt(int64(len(files)), 10))
					for fi, file := range files {
						fiStr := strconv.FormatInt(int64(fi), 10)
						ui.Machine("artifact", iStr, "file", fiStr, file)
					}
				} else {
					ui.Machine("artifact", iStr, "nil")
				}

				ui.Machine("artifact", iStr, "end")
				c.Ui.Say(message.String())
			}
		}
	} else {
		c.Ui.Say("\n==> Builds finished but no artifacts were created.")
	}

	if len(errors) > 0 {
		// If any errors occurred, exit with a non-zero exit status
		return 1
	}

	return 0
}

func (BuildCommand) Help() string {
	helpText := `
Usage: packer build [options] TEMPLATE

  Will execute multiple builds in parallel as defined in the template.
  The various artifacts created by the template will be outputted.

Options:

  -color=false               Disable color output (on by default)
  -debug                     Debug mode enabled for builds
  -except=foo,bar,baz        Build all builds other than these
  -force                     Force a build to continue if artifacts exist, deletes existing artifacts
  -machine-readable          Machine-readable output
  -only=foo,bar,baz          Only build the given builds by name
  -parallel=false            Disable parallelization (on by default)
  -var 'key=value'           Variable for templates, can be used multiple times.
  -var-file=path             JSON file containing user variables.
`

	return strings.TrimSpace(helpText)
}

func (BuildCommand) Synopsis() string {
	return "build image(s) from template"
}
