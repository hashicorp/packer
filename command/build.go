// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/internal/hcp/registry"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/sync/semaphore"

	"github.com/hako/durafmt"
	"github.com/posener/complete"
)

const (
	hcpReadyIntegrationURL = "https://developer.hashicorp.com/packer/integrations?flags=hcp-ready"
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
	flags := c.Meta.FlagSet("build")
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

func (c *BuildCommand) RunContext(buildCtx context.Context, cla *BuildArgs) int {
	// Set the release only flag if specified as argument
	//
	// This deactivates the capacity for Packer to load development binaries.
	c.CoreConfig.Components.PluginConfig.ReleasesOnly = cla.ReleaseOnly

	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	diags := packerStarter.DetectPluginBinaries()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	diags = packerStarter.Initialize(packer.InitializeOptions{})
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	hcpRegistry, diags := registry.New(packerStarter, c.Ui)
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	defer hcpRegistry.VersionStatusSummary()

	err := hcpRegistry.PopulateVersion(buildCtx)
	if err != nil {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary:  "HCP: populating version failed",
				Severity: hcl.DiagError,
				Detail:   err.Error(),
			},
		})
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
	if len(builds) == 0 && ret != 0 {
		return ret
	}

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
	buildUis := make(map[packersdk.Build]packersdk.Ui)
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

	if len(builds) == 0 {
		return writeDiags(c.Ui, nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "No builds to run",
				Detail: "A build command cannot run without at least one build to process. " +
					"If the only or except flags have been specified at run time check that" +
					" at least one build is selected for execution.",
				Severity: hcl.DiagError,
			},
		})
	}

	// Get the start of the build command
	buildCommandStart := time.Now()

	// Run all the builds in parallel and wait for them to complete
	var wg sync.WaitGroup
	var artifacts = struct {
		sync.RWMutex
		m map[string][]packersdk.Artifact
	}{m: make(map[string][]packersdk.Artifact)}
	// Get the builds we care about
	var errs = struct {
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
			errs.Lock()
			errs.m[name] = err
			errs.Unlock()
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

			err := hcpRegistry.StartBuild(buildCtx, b)
			// Seems odd to require this error check here. Now that it is an error we can just exit with diag
			if err != nil {
				// If the build is already done, we skip without a warning
				if errors.As(err, &registry.ErrBuildAlreadyDone{}) {
					ui.Say(fmt.Sprintf("skipping already done build %q", name))
					return
				}
				writeDiags(c.Ui, nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Summary: fmt.Sprintf(
							"hcp: failed to start build %q",
							name),
						Severity: hcl.DiagError,
						Detail:   err.Error(),
					},
				})
				return
			}

			log.Printf("Starting build run: %s", name)
			runArtifacts, err := b.Run(buildCtx, ui)

			// Get the duration of the build and parse it
			buildEnd := time.Now()
			buildDuration := buildEnd.Sub(buildStart)
			fmtBuildDuration := durafmt.Parse(buildDuration).LimitFirstN(2)

			runArtifacts, hcperr := hcpRegistry.CompleteBuild(
				buildCtx,
				b,
				runArtifacts,
				err)
			if hcperr != nil {
				if _, ok := hcperr.(*registry.NotAHCPArtifactError); ok {
					writeDiags(c.Ui, nil, hcl.Diagnostics{
						&hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  fmt.Sprintf("The %q builder produced an artifact that cannot be pushed to HCP Packer", b.Name()),
							Detail: fmt.Sprintf(
								`%s
Check that you are using an HCP Ready integration before trying again:
%s`,
								hcperr, hcpReadyIntegrationURL),
						},
					})
				} else {
					writeDiags(c.Ui, nil, hcl.Diagnostics{
						&hcl.Diagnostic{
							Summary: fmt.Sprintf(
								"publishing build metadata to HCP Packer for %q failed",
								name),
							Severity: hcl.DiagError,
							Detail:   hcperr.Error(),
						},
					})
				}
			}

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored after %s: %s", name, fmtBuildDuration, err))
				errs.Lock()
				errs.m[name] = err
				errs.Unlock()
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished after %s.", name, fmtBuildDuration))
				if runArtifacts != nil {
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

	if len(errs.m) > 0 {
		c.Ui.Machine("error-count", strconv.FormatInt(int64(len(errs.m)), 10))

		c.Ui.Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errs.m {
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

	if len(errs.m) > 0 {
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
  -var-file=path                JSON or HCL2 file containing user variables, can be used multiple times.
  -warn-on-undeclared-var       Display warnings for user variable files containing undeclared variables.
  -ignore-prerelease-plugins    Disable the loading of prerelease plugin binaries (x.y.z-dev).
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
