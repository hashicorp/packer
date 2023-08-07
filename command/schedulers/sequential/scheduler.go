package sequential

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hako/durafmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/internal/hcp/registry"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/sync/semaphore"

	opts "github.com/hashicorp/packer/command/schedulers/opts"
	hclscheduler "github.com/hashicorp/packer/command/schedulers/sequential/hcl"
	jsonscheduler "github.com/hashicorp/packer/command/schedulers/sequential/json"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type SpecialisedSequentialScheduler interface {
	EvaluateDataSources() hcl.Diagnostics
	EvaluateVariables() hcl.Diagnostics
	EvaluateBuilds() hcl.Diagnostics
	GetBuilds() ([]packersdk.Build, hcl.Diagnostics)
	Options() *opts.SchedulerOptions
}

type SequentialScheduler struct {
	scheduler SpecialisedSequentialScheduler
	ui        packersdk.Ui
	handler   packer.Handler
	RunBuilds bool
	context   context.Context

	useHCP      bool
	hcpRegistry registry.Registry
}

func NewSequentialScheduler(h packer.Handler, opts opts.SchedulerOptions) *SequentialScheduler {
	sched := &SequentialScheduler{
		handler: h,
	}

	switch handler := h.(type) {
	case *packer.Core:
		sched.scheduler = jsonscheduler.NewScheduler(handler, &opts)
	case *hcl2template.PackerConfig:
		sched.scheduler = hclscheduler.NewScheduler(handler, &opts)
	}

	return sched
}

func (s *SequentialScheduler) WithBuilds() *SequentialScheduler {
	s.RunBuilds = true
	return s
}

func (s *SequentialScheduler) WithHCPRegistry() *SequentialScheduler {
	s.useHCP = true
	return s
}

func (s *SequentialScheduler) WithSkipDatasourceExecution() *SequentialScheduler {
	s.scheduler.Options().SkipDatasourcesExecution = true
	return s
}

func (s *SequentialScheduler) WithContext(ctx context.Context) *SequentialScheduler {
	s.context = ctx
	return s
}

func (s *SequentialScheduler) WithUi(ui packersdk.Ui) *SequentialScheduler {
	s.ui = ui
	return s
}

func (s *SequentialScheduler) Run() hcl.Diagnostics {
	var diags hcl.Diagnostics

	diags = append(diags, s.scheduler.EvaluateDataSources()...)
	if diags.HasErrors() {
		return diags
	}

	diags = append(diags, s.scheduler.EvaluateVariables()...)
	if diags.HasErrors() {
		return diags
	}

	diags = append(diags, s.scheduler.EvaluateBuilds()...)
	if diags.HasErrors() {
		return diags
	}

	if s.useHCP {
		s.hcpRegistry, diags = registry.New(s.handler, s.ui)
		if diags.HasErrors() {
			return diags
		}
	} else {
		s.hcpRegistry = registry.NewNullRegistry()
	}

	defer s.hcpRegistry.IterationStatusSummary()

	err := s.hcpRegistry.PopulateIteration(s.context)
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary:  "HCP: populating iteration failed",
				Severity: hcl.DiagError,
				Detail:   err.Error(),
			},
		}
	}

	builds, buildDiags := s.scheduler.GetBuilds()
	diags = append(diags, buildDiags...)
	if diags.HasErrors() {
		return diags
	}

	if !s.RunBuilds {
		return diags
	}

	s.runBuilds(builds)

	return nil
}

func (s *SequentialScheduler) runBuilds(builds []packersdk.Build) hcl.Diagnostics {
	if s.scheduler.Options().Debug {
		s.ui.Say("Debug mode enabled. Builds will not be parallelized.")
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
		ui := s.ui
		if s.scheduler.Options().Color {
			// Only set up UI colors if -machine-readable isn't set.
			if _, ok := s.ui.(*packer.MachineReadableUi); !ok {
				ui = &packer.ColoredUi{
					Color: colors[i%len(colors)],
					Ui:    ui,
				}
				ui.Say(fmt.Sprintf("%s: output will be in this color.", builds[i].Name()))
				if i+1 == len(builds) {
					// Add a newline between the color output and the actual output
					s.ui.Say("")
				}
			}
		}
		// Now add timestamps if requested
		if s.scheduler.Options().TimestampUi {
			ui = &packer.TimestampedUi{
				Ui: ui,
			}
		}

		buildUis[builds[i]] = ui
	}
	log.Printf("Build debug mode: %v", s.scheduler.Options().Debug)
	log.Printf("Force build: %v", s.scheduler.Options().Force)
	log.Printf("On error: %v", s.scheduler.Options().OnError)

	if len(builds) == 0 {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "No builds to run",
				Detail: "A build command cannot run without at least one build to process. " +
					"If the only or except flags have been specified at run time check that" +
					" at least one build is selected for execution.",
				Severity: hcl.DiagError,
			},
		}
	}

	var diags hcl.Diagnostics

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
	limitParallel := semaphore.NewWeighted(s.scheduler.Options().ParallelBuilds)
	for i := range builds {
		if err := s.context.Err(); err != nil {
			log.Println("Interrupted, not going to start any more builds.")
			break
		}

		b := builds[i]
		name := b.Name()
		ui := buildUis[b]
		if err := limitParallel.Acquire(s.context, 1); err != nil {
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

			err := s.hcpRegistry.StartBuild(s.context, b)
			// Seems odd to require this error check here. Now that it is an error we can just exit with diag
			if err != nil {
				// If the build is already done, we skip without a warning
				if errors.As(err, &registry.ErrBuildAlreadyDone{}) {
					ui.Say(fmt.Sprintf("skipping already done build %q", name))
					return
				}
				diags = append(diags, &hcl.Diagnostic{
					Summary: fmt.Sprintf(
						"hcp: failed to start build %q",
						name),
					Severity: hcl.DiagError,
					Detail:   err.Error(),
				})
				return
			}

			log.Printf("Starting build run: %s", name)
			runArtifacts, err := b.Run(s.context, ui)

			// Get the duration of the build and parse it
			buildEnd := time.Now()
			buildDuration := buildEnd.Sub(buildStart)
			fmtBuildDuration := durafmt.Parse(buildDuration).LimitFirstN(2)

			runArtifacts, hcperr := s.hcpRegistry.CompleteBuild(
				s.context,
				b,
				runArtifacts,
				err)
			if hcperr != nil {
				diags = append(diags, &hcl.Diagnostic{
					Summary: fmt.Sprintf(
						"failed to complete HCP-enabled build %q",
						name),
					Severity: hcl.DiagError,
					Detail:   hcperr.Error(),
				})
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

		if s.scheduler.Options().Debug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if s.scheduler.Options().ParallelBuilds == 1 {
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
	s.ui.Say(fmt.Sprintf("\n==> Wait completed after %s", fmtBuildCommandDuration))

	if err := s.context.Err(); err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Build cancelled",
				Detail:   "Cleanly cancelled builds after being interrupted.",
			},
		}
	}

	if len(errs.m) > 0 {
		s.ui.Machine("error-count", strconv.FormatInt(int64(len(errs.m)), 10))

		s.ui.Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errs.m {
			// Create a UI for the machine readable stuff to be targeted
			ui := &packer.TargetedUI{
				Target: name,
				Ui:     s.ui,
			}

			ui.Machine("error", err.Error())

			s.ui.Error(fmt.Sprintf("--> %s: %s", name, err))
		}
	}

	if len(artifacts.m) > 0 {
		s.ui.Say("\n==> Builds finished. The artifacts of successful builds are:")
		for name, buildArtifacts := range artifacts.m {
			// Create a UI for the machine readable stuff to be targeted
			ui := &packer.TargetedUI{
				Target: name,
				Ui:     s.ui,
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
				s.ui.Say(message.String())

			}

		}
	} else {
		s.ui.Say("\n==> Builds finished but no artifacts were created.")
	}

	return diags
}
