package command

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/helper/enumflag"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template"
	"golang.org/x/sync/semaphore"

	"github.com/posener/complete"
)

type BuildCommand struct {
	Meta
}

func (c *BuildCommand) Run(args []string) int {
	buildCtx, cancelBuildCtx := context.WithCancel(context.Background())
	// Handle interrupts for this build
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer func() {
		cancelBuildCtx()
		signal.Stop(sigCh)
		close(sigCh)
	}()
	go func() {
		select {
		case sig := <-sigCh:
			if sig == nil {
				// context got cancelled and this closed chan probably
				// triggered first
				return
			}
			c.Ui.Error(fmt.Sprintf("Cancelling build after receiving %s", sig))
			cancelBuildCtx()
		case <-buildCtx.Done():
		}
	}()

	return c.RunContext(buildCtx, args)
}

// Config is the command-configuration parsed from the command line.
type Config struct {
	Color, Debug, Force, Timestamp bool
	ParallelBuilds                 int64
	OnError                        string
	Path                           string
}

func (c *BuildCommand) ParseArgs(args []string) (Config, int) {
	var cfg Config
	var parallel bool
	flags := c.Meta.FlagSet("build", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	flags.BoolVar(&cfg.Color, "color", true, "")
	flags.BoolVar(&cfg.Debug, "debug", false, "")
	flags.BoolVar(&cfg.Force, "force", false, "")
	flags.BoolVar(&cfg.Timestamp, "timestamp-ui", false, "")
	flagOnError := enumflag.New(&cfg.OnError, "cleanup", "abort", "ask")
	flags.Var(flagOnError, "on-error", "")
	flags.BoolVar(&parallel, "parallel", true, "")
	flags.Int64Var(&cfg.ParallelBuilds, "parallel-builds", 0, "")
	if err := flags.Parse(args); err != nil {
		return cfg, 1
	}

	if parallel == false && cfg.ParallelBuilds == 0 {
		cfg.ParallelBuilds = 1
	}
	if cfg.ParallelBuilds < 1 {
		cfg.ParallelBuilds = math.MaxInt64
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return cfg, 1
	}
	cfg.Path = args[0]
	return cfg, 0
}

func (c *BuildCommand) GetBuildsFromHCL(path string) ([]packer.Build, int) {
	parser := &hcl2template.Parser{
		Parser:                hclparse.NewParser(),
		BuilderSchemas:        c.CoreConfig.Components.BuilderStore,
		ProvisionersSchemas:   c.CoreConfig.Components.ProvisionerStore,
		PostProcessorsSchemas: c.CoreConfig.Components.PostProcessorStore,
	}

	builds, diags := parser.Parse(path, c.flagVars)
	{
		// write HCL errors/diagnostics if any.
		b := bytes.NewBuffer(nil)
		err := hcl.NewDiagnosticTextWriter(b, parser.Files(), 80, false).WriteDiagnostics(diags)
		if err != nil {
			c.Ui.Error("could not write diagnostic: " + err.Error())
			return nil, 1
		}
		if b.Len() != 0 {
			c.Ui.Message(b.String())
		}
	}
	ret := 0
	if diags.HasErrors() {
		ret = 1
	}

	return builds, ret
}

func (c *BuildCommand) GetBuilds(path string) ([]packer.Build, int) {

	isHCLLoaded, err := isHCLLoaded(path)
	if path != "-" && err != nil {
		c.Ui.Error(fmt.Sprintf("could not tell whether %s is hcl enabled: %s", path, err))
		return nil, 1
	}
	if isHCLLoaded {
		return c.GetBuildsFromHCL(path)
	}

	// TODO: uncomment in v1.5.1 once we've polished HCL a bit more.
	// c.Ui.Say(`Legacy JSON Configuration Will Be Used.
	// The template will be parsed in the legacy configuration style. This style
	// will continue to work but users are encouraged to move to the new style.
	// See: https://packer.io/guides/hcl
	// `)

	// Parse the template
	var tpl *template.Template
	tpl, err = template.ParseFile(path)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return nil, 1
	}

	// Get the core
	core, err := c.Meta.Core(tpl)
	if err != nil {
		c.Ui.Error(err.Error())
		return nil, 1
	}

	ret := 0
	buildNames := c.Meta.BuildNames(core)
	builds := make([]packer.Build, 0, len(buildNames))
	for _, n := range buildNames {
		b, err := core.Build(n)
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to initialize build '%s': %s",
				n, err))
			ret = 1
			continue
		}

		builds = append(builds, b)
	}
	return builds, ret
}

func (c *BuildCommand) RunContext(buildCtx context.Context, args []string) int {
	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	builds, ret := c.GetBuilds(cfg.Path)

	if cfg.Debug {
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
	buildUis := make(map[packer.Build]packer.Ui)
	for i := range builds {
		ui := c.Ui
		if cfg.Color {
			ui = &packer.ColoredUi{
				Color: colors[i%len(colors)],
				Ui:    ui,
			}
			if _, ok := c.Ui.(*packer.MachineReadableUi); !ok {
				ui.Say(fmt.Sprintf("%s: output will be in this color.", builds[i].Name()))
				if i+1 == len(builds) {
					// Add a newline between the color output and the actual output
					c.Ui.Say("")
				}
			}
		}
		// Now add timestamps if requested
		if cfg.Timestamp {
			ui = &packer.TimestampedUi{
				Ui: ui,
			}
		}

		buildUis[builds[i]] = ui
	}

	log.Printf("Build debug mode: %v", cfg.Debug)
	log.Printf("Force build: %v", cfg.Force)
	log.Printf("On error: %v", cfg.OnError)

	// Set the debug and force mode and prepare all the builds
	for i := range builds {
		b := builds[i]
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(cfg.Debug)
		b.SetForce(cfg.Force)
		b.SetOnError(cfg.OnError)

		warnings, err := b.Prepare()
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		if len(warnings) > 0 {
			ui := buildUis[b]
			ui.Say(fmt.Sprintf("Warnings for build '%s':\n", b.Name()))
			for _, warning := range warnings {
				ui.Say(fmt.Sprintf("* %s", warning))
			}
			ui.Say("")
		}
	}

	// Run all the builds in parallel and wait for them to complete
	var wg sync.WaitGroup
	var artifacts = struct {
		sync.RWMutex
		m map[string][]packer.Artifact
	}{m: make(map[string][]packer.Artifact)}
	// Get the builds we care about
	var errors = struct {
		sync.RWMutex
		m map[string]error
	}{m: make(map[string]error)}
	limitParallel := semaphore.NewWeighted(cfg.ParallelBuilds)
	for i := range builds {
		if err := buildCtx.Err(); err != nil {
			log.Println("Interrupted, not going to start any more builds.")
			break
		}

		b := builds[i]
		name := b.Name()
		ui := buildUis[b]
		if err := limitParallel.Acquire(buildCtx, 1); err != nil {
			ui.Error(fmt.Sprintf("Build '%s' failed to acquire semaphore: %s", name, err))
			errors.Lock()
			errors.m[name] = err
			errors.Unlock()
			break
		}
		// Increment the waitgroup so we wait for this item to finish properly
		wg.Add(1)

		// Run the build in a goroutine
		go func() {
			defer wg.Done()

			defer limitParallel.Release(1)

			log.Printf("Starting build run: %s", name)
			runArtifacts, err := b.Run(buildCtx, ui)

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored: %s", name, err))
				errors.Lock()
				errors.m[name] = err
				errors.Unlock()
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished.", name))
				if nil != runArtifacts {
					artifacts.Lock()
					artifacts.m[name] = runArtifacts
					artifacts.Unlock()
				}
			}
		}()

		if cfg.Debug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if cfg.ParallelBuilds == 1 {
			log.Printf("Parallelization disabled, waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

	}

	// Wait for both the builds to complete and the interrupt handler,
	// if it is interrupted.
	log.Printf("Waiting on builds to complete...")
	wg.Wait()

	if err := buildCtx.Err(); err != nil {
		c.Ui.Say("Cleanly cancelled builds after being interrupted.")
		return 1
	}

	if len(errors.m) > 0 {
		c.Ui.Machine("error-count", strconv.FormatInt(int64(len(errors.m)), 10))

		c.Ui.Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errors.m {
			// Create a UI for the machine readable stuff to be targeted
			ui := &packer.TargetedUI{
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
			// Create a UI for the machine readable stuff to be targeted
			ui := &packer.TargetedUI{
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

	if len(errors.m) > 0 {
		// If any errors occurred, exit with a non-zero exit status
		ret = 1
	}

	return ret
}

func (*BuildCommand) Help() string {
	helpText := `
Usage: packer build [options] TEMPLATE

  Will execute multiple builds in parallel as defined in the template.
  The various artifacts created by the template will be outputted.

Options:

  -color=false                  Disable color output. (Default: color)
  -debug                        Debug mode enabled for builds.
  -except=foo,bar,baz           Run all builds and post-procesors other than these.
  -only=foo,bar,baz             Build only the specified builds.
  -force                        Force a build to continue if artifacts exist, deletes existing artifacts.
  -machine-readable             Produce machine-readable output.
  -on-error=[cleanup|abort|ask] If the build fails do: clean up (default), abort, or ask.
  -parallel=false               Disable parallelization. (Default: true)
  -parallel-builds=1            Number of builds to run in parallel. 0 means no limit (Default: 0)
  -timestamp-ui                 Enable prefixing of each ui output with an RFC3339 timestamp.
  -var 'key=value'              Variable for templates, can be used multiple times.
  -var-file=path                JSON file containing user variables. [ Note that even in HCL mode this expects file to contain JSON, a fix is comming soon ]
`

	return strings.TrimSpace(helpText)
}

func (*BuildCommand) Synopsis() string {
	return "build image(s) from template"
}

func (*BuildCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*BuildCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-color":            complete.PredictNothing,
		"-debug":            complete.PredictNothing,
		"-except":           complete.PredictNothing,
		"-only":             complete.PredictNothing,
		"-force":            complete.PredictNothing,
		"-machine-readable": complete.PredictNothing,
		"-on-error":         complete.PredictNothing,
		"-parallel":         complete.PredictNothing,
		"-timestamp-ui":     complete.PredictNothing,
		"-var":              complete.PredictNothing,
		"-var-file":         complete.PredictNothing,
	}
}
