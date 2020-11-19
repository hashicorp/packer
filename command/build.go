package command

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template"
	"github.com/hashicorp/packer/version"
	"golang.org/x/sync/semaphore"

	"github.com/hako/durafmt"
	"github.com/posener/complete"
)

type BuildCommand struct {
	Meta
}

func (c *BuildCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *BuildCommand) ParseArgs(args []string) (*BuildArgs, int) {
	var cfg BuildArgs
	flags := c.Meta.FlagSet("build", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	if cfg.ParallelBuilds < 1 {
		cfg.ParallelBuilds = math.MaxInt64
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.Path = args[0]
	return &cfg, 0
}

func (m *Meta) GetConfigFromHCL(cla *MetaArgs) (*hcl2template.PackerConfig, int) {
	parser := &hcl2template.Parser{
		CorePackerVersion:       version.SemVer,
		CorePackerVersionString: version.FormattedVersion(),
		Parser:                  hclparse.NewParser(),
		BuilderSchemas:          m.CoreConfig.Components.BuilderStore,
		ProvisionersSchemas:     m.CoreConfig.Components.ProvisionerStore,
		PostProcessorsSchemas:   m.CoreConfig.Components.PostProcessorStore,
	}
	cfg, diags := parser.Parse(cla.Path, cla.VarFiles, cla.Vars)
	return cfg, writeDiags(m.Ui, parser.Files(), diags)
}

func writeDiags(ui packersdk.Ui, files map[string]*hcl.File, diags hcl.Diagnostics) int {
	// write HCL errors/diagnostics if any.
	b := bytes.NewBuffer(nil)
	err := hcl.NewDiagnosticTextWriter(b, files, 80, false).WriteDiagnostics(diags)
	if err != nil {
		ui.Error("could not write diagnostic: " + err.Error())
		return 1
	}
	if b.Len() != 0 {
		if diags.HasErrors() {
			ui.Error(b.String())
			return 1
		}
		ui.Say(b.String())
	}
	return 0
}

func (m *Meta) GetConfig(cla *MetaArgs) (packer.Handler, int) {
	cfgType, err := cla.GetConfigType()
	if err != nil {
		m.Ui.Error(fmt.Sprintf("%q: %s", cla.Path, err))
		return nil, 1
	}

	switch cfgType {
	case ConfigTypeHCL2:
		// TODO(azr): allow to pass a slice of files here.
		return m.GetConfigFromHCL(cla)
	default:
		// TODO: uncomment once we've polished HCL a bit more.
		// c.Ui.Say(`Legacy JSON Configuration Will Be Used.
		// The template will be parsed in the legacy configuration style. This style
		// will continue to work but users are encouraged to move to the new style.
		// See: https://packer.io/guides/hcl
		// `)
		return m.GetConfigFromJSON(cla)
	}
}

func (m *Meta) GetConfigFromJSON(cla *MetaArgs) (packer.Handler, int) {
	// Parse the template
	var tpl *template.Template
	var err error
	if cla.Path == "" {
		// here cla validation passed so this means we want a default builder
		// and we probably are in the console command
		tpl, err = template.Parse(TiniestBuilder)
	} else {
		tpl, err = template.ParseFile(cla.Path)
	}

	if err != nil {
		m.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return nil, 1
	}

	// Get the core
	core, err := m.Core(tpl, cla)
	ret := 0
	if err != nil {
		m.Ui.Error(err.Error())
		ret = 1
	}
	return &CoreWrapper{core}, ret
}

func (c *BuildCommand) RunContext(buildCtx context.Context, cla *BuildArgs) int {
	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}
	diags := packerStarter.Initialize()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	builds, diags := packerStarter.GetBuilds(packer.GetBuildsOptions{
		Only:    cla.Only,
		Except:  cla.Except,
		Debug:   cla.Debug,
		Force:   cla.Force,
		OnError: cla.OnError,
	})

	// here, something could have gone wrong but we still want to run valid
	// builds.
	ret = writeDiags(c.Ui, nil, diags)

	if cla.Debug {
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
	buildUis := make(map[packer.Build]packersdk.Ui)
	for i := range builds {
		ui := c.Ui
		if cla.Color {
			// Only set up UI colors if -machine-readable isn't set.
			if _, ok := c.Ui.(*packer.MachineReadableUi); !ok {
				ui = &packer.ColoredUi{
					Color: colors[i%len(colors)],
					Ui:    ui,
				}
				ui.Say(fmt.Sprintf("%s: output will be in this color.", builds[i].Name()))
				if i+1 == len(builds) {
					// Add a newline between the color output and the actual output
					c.Ui.Say("")
				}
			}
		}
		// Now add timestamps if requested
		if cla.TimestampUi {
			ui = &packer.TimestampedUi{
				Ui: ui,
			}
		}

		buildUis[builds[i]] = ui
	}

	log.Printf("Build debug mode: %v", cla.Debug)
	log.Printf("Force build: %v", cla.Force)
	log.Printf("On error: %v", cla.OnError)

	// Get the start of the build command
	buildCommandStart := time.Now()

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
	limitParallel := semaphore.NewWeighted(cla.ParallelBuilds)
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
			// Get the start of the build
			buildStart := time.Now()

			defer wg.Done()

			defer limitParallel.Release(1)

			log.Printf("Starting build run: %s", name)
			runArtifacts, err := b.Run(buildCtx, ui)

			// Get the duration of the build and parse it
			buildEnd := time.Now()
			buildDuration := buildEnd.Sub(buildStart)
			fmtBuildDuration := durafmt.Parse(buildDuration).LimitFirstN(2)

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored after %s: %s", name, fmtBuildDuration, err))
				errors.Lock()
				errors.m[name] = err
				errors.Unlock()
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished after %s.", name, fmtBuildDuration))
				if nil != runArtifacts {
					artifacts.Lock()
					artifacts.m[name] = runArtifacts
					artifacts.Unlock()
				}
			}
		}()

		if cla.Debug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if cla.ParallelBuilds == 1 {
			log.Printf("Parallelization disabled, waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

	}

	// Wait for both the builds to complete and the interrupt handler,
	// if it is interrupted.
	log.Printf("Waiting on builds to complete...")
	wg.Wait()

	// Get the duration of the buildCommand command and parse it
	buildCommandEnd := time.Now()
	buildCommandDuration := buildCommandEnd.Sub(buildCommandStart)
	fmtBuildCommandDuration := durafmt.Parse(buildCommandDuration).LimitFirstN(2)
	c.Ui.Say(fmt.Sprintf("\n==> Wait completed after %s", fmtBuildCommandDuration))

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
  -except=foo,bar,baz           Run all builds and post-processors other than these.
  -only=foo,bar,baz             Build only the specified builds.
  -force                        Force a build to continue if artifacts exist, deletes existing artifacts.
  -machine-readable             Produce machine-readable output.
  -on-error=[cleanup|abort|ask|run-cleanup-provisioner] If the build fails do: clean up (default), abort, ask, or run-cleanup-provisioner.
  -parallel-builds=1            Number of builds to run in parallel. 1 disables parallelization. 0 means no limit (Default: 0)
  -timestamp-ui                 Enable prefixing of each ui output with an RFC3339 timestamp.
  -var 'key=value'              Variable for templates, can be used multiple times.
  -var-file=path                JSON or HCL2 file containing user variables.
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
