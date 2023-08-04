package schedulers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/hako/durafmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/didyoumean"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/hcl2template"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/internal/hcp/registry"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/sync/semaphore"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	ttmp "text/template"
)

type SequentialScheduler struct {
	ui                       packersdk.Ui
	handler                  packer.Handler
	skipDatasourcesExecution bool
	RunBuilds                bool
	context                  context.Context

	useHCP      bool
	hcpRegistry registry.Registry

	options SchedulerOptions
}

func NewSequentialScheduler(h packer.Handler, opts SchedulerOptions) *SequentialScheduler {
	return &SequentialScheduler{
		handler: h,
		options: opts,
	}
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
	s.skipDatasourcesExecution = true
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

	diags = append(diags, s.EvaluateDataSources()...)
	if diags.HasErrors() {
		return diags
	}

	diags = append(diags, s.EvaluateVariables()...)
	if diags.HasErrors() {
		return diags
	}

	diags = append(diags, s.EvaluateBuilds()...)
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

	builds, buildDiags := s.GetBuilds()
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

func (s *SequentialScheduler) EvaluateBuilds() hcl.Diagnostics {
	switch s.handler.(type) {
	case *packer.Core:
		return nil
	case *hcl2template.PackerConfig:
		return s.hcl2EvaluateBuilds()
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown configuration type",
			Detail: fmt.Sprintf(`
The packer handler is of unknown type %q, expected either a *packer.Core or a *hcl2template.PackerConfig

This is likely a Packer bug, please report this so the team can take a look at it.`,
				reflect.TypeOf(s.handler).String()),
		},
	}
}

func (s *SequentialScheduler) hcl2EvaluateBuilds() hcl.Diagnostics {
	cfg := s.handler.(*hcl2template.PackerConfig)

	var diags hcl.Diagnostics

	// parse the actual content // rest
	for _, file := range cfg.Files {
		diags = append(diags, cfg.Parser.ParseConfig(file, cfg)...)
	}

	diags = append(diags, initializeBlocks(cfg)...)

	return diags
}

func initializeBlocks(cfg *hcl2template.PackerConfig) hcl.Diagnostics {
	// verify that all used plugins do exist
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for i := range build.Sources {
			// here we grab a pointer to the source usage because we will set
			// its body.
			srcUsage := &(build.Sources[i])
			if !cfg.Parser.PluginConfig.Builders.Has(srcUsage.Type) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + hcl2template.BuildSourceLabel + " type " + srcUsage.Type,
					Subject:  &build.HCL2Ref.DefRange,
					Detail:   fmt.Sprintf("known builders: %v", cfg.Parser.PluginConfig.Builders.List()),
					Severity: hcl.DiagError,
				})
				continue
			}

			sourceDefinition, found := cfg.Sources[srcUsage.SourceRef]
			if !found {
				availableSrcs := hcl2template.ListAvailableSourceNames(cfg.Sources)
				detail := fmt.Sprintf("Known: %v", availableSrcs)
				if sugg := didyoumean.NameSuggestion(srcUsage.SourceRef.String(), availableSrcs); sugg != "" {
					detail = fmt.Sprintf("Did you mean to use %q?", sugg)
				}
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + hcl2template.SourceLabel + " " + srcUsage.SourceRef.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   detail,
				})
				continue
			}

			body := sourceDefinition.Block.Body
			if srcUsage.Body != nil {
				// merge additions into source definition to get a new body.
				body = hcl.MergeBodies([]hcl.Body{body, srcUsage.Body})
			}

			srcUsage.Body = body
		}

		for _, provBlock := range build.ProvisionerBlocks {
			if !cfg.Parser.PluginConfig.Provisioners.Has(provBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildProvisionerLabel+" type %q", provBlock.PType),
					Subject:  provBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+hcl2template.BuildProvisionerLabel+"s: %v", cfg.Parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
		}

		if build.ErrorCleanupProvisionerBlock != nil {
			if !cfg.Parser.PluginConfig.Provisioners.Has(build.ErrorCleanupProvisionerBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildErrorCleanupProvisionerLabel+" type %q", build.ErrorCleanupProvisionerBlock.PType),
					Subject:  build.ErrorCleanupProvisionerBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+hcl2template.BuildErrorCleanupProvisionerLabel+"s: %v", cfg.Parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
		}

		for _, ppList := range build.PostProcessorsLists {
			for _, ppBlock := range ppList {
				if !cfg.Parser.PluginConfig.PostProcessors.Has(ppBlock.PType) {
					diags = append(diags, &hcl.Diagnostic{
						Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildPostProcessorLabel+" type %q", ppBlock.PType),
						Subject:  ppBlock.HCL2Ref.TypeRange.Ptr(),
						Detail:   fmt.Sprintf("known "+hcl2template.BuildPostProcessorLabel+"s: %v", cfg.Parser.PluginConfig.PostProcessors.List()),
						Severity: hcl.DiagError,
					})
				}
			}
		}

	}

	return diags
}

func (s *SequentialScheduler) runBuilds(builds []packersdk.Build) hcl.Diagnostics {
	if s.options.Debug {
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
		if s.options.Color {
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
		if s.options.TimestampUi {
			ui = &packer.TimestampedUi{
				Ui: ui,
			}
		}

		buildUis[builds[i]] = ui
	}
	log.Printf("Build debug mode: %v", s.options.Debug)
	log.Printf("Force build: %v", s.options.Force)
	log.Printf("On error: %v", s.options.OnError)

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
	limitParallel := semaphore.NewWeighted(s.options.ParallelBuilds)
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

		if s.options.Debug {
			log.Printf("Debug enabled, so waiting for build to finish: %s", b.Name())
			wg.Wait()
		}

		if s.options.ParallelBuilds == 1 {
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

func (s *SequentialScheduler) EvaluateVariables() hcl.Diagnostics {
	switch s.handler.(type) {
	case *packer.Core:
		return s.jsonVariableEval()
	case *hcl2template.PackerConfig:
		return s.hcl2EvaluateLocalVariables()
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown configuration type",
			Detail: `
The packer handler is of unknown type %q, expected either a *packer.Core or a *hcl2template.PackerConfig

This is likely a Packer bug, please report this so the team can take a look at it.`,
		},
	}
}

func (s *SequentialScheduler) GetBuilds() ([]packersdk.Build, hcl.Diagnostics) {
	switch s.handler.(type) {
	case *packer.Core:
		return s.jsonGetBuilds(s.options.toPackerBuildOpts())
	case *hcl2template.PackerConfig:
		return s.hcl2GetBuilds(s.options.toPackerBuildOpts())
	}

	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown configuration type",
			Detail: `
The packer handler is of unknown type %q, expected either a *packer.Core or a *hcl2template.PackerConfig

This is likely a Packer bug, please report this so the team can take a look at it.`,
		},
	}
}

// Convert -only and -except globs to glob.Glob instances.
func convertFilterOption(patterns []string, optionName string) ([]glob.Glob, hcl.Diagnostics) {
	var globs []glob.Glob
	var diags hcl.Diagnostics

	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  fmt.Sprintf("Invalid -%s pattern %s: %s", optionName, pattern, err),
				Severity: hcl.DiagError,
			})
		}
		globs = append(globs, g)
	}

	return globs, diags
}

func warningErrorsToDiags(block *hcl.Block, warnings []string, err error) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, warning := range warnings {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  warning,
			Subject:  &block.DefRange,
			Severity: hcl.DiagWarning,
		})
	}
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}
	return diags
}

func startBuilder(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, ectx *hcl.EvalContext) (packersdk.Builder, hcl.Diagnostics, []string) {
	var diags hcl.Diagnostics

	builder, err := cfg.Parser.PluginConfig.Builders.Start(source.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to load " + hcl2template.SourceLabel + " type",
			Detail:   err.Error(),
		})
		return builder, diags, nil
	}

	body := source.Body
	// Add known values to source accessor in eval context.
	ectx.Variables[hcl2template.SourcesAccessor] = cty.ObjectVal(source.CtyValues())

	decoded, moreDiags := hcl2template.DecodeHCL2Spec(body, ectx, builder)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return builder, diags, nil
	}

	// In case of cty.Unknown values, this will write a equivalent placeholder of the same type
	// Unknown types are not recognized by the json marshal during the RPC call and we have to do this here
	// to avoid json parsing failures when running the validate command.
	// We don't do this before so we can validate if variable types matches correctly on decodeHCL2Spec.
	decoded = hcl2shim.WriteUnknownPlaceholderValues(decoded)

	// Note: HCL prepares inside of the Start func, but Json does not. Json
	// builds are instead prepared only in command/build.go
	// TODO: either make json prepare when plugins are loaded, or make HCL
	// prepare at a later step, to make builds from different template types
	// easier to reason about.
	builderVars := source.BuilderVariables()
	builderVars["packer_core_version"] = cfg.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(cfg.Debug)
	builderVars["packer_force"] = strconv.FormatBool(cfg.Force)
	builderVars["packer_on_error"] = cfg.OnError

	generatedVars, warning, err := builder.Prepare(builderVars, decoded)
	moreDiags = warningErrorsToDiags(cfg.Sources[source.SourceRef].Block, warning, err)
	diags = append(diags, moreDiags...)
	return builder, diags, generatedVars
}

// getCoreBuildProvisioners takes a list of provisioner block, starts according
// provisioners and sends parsed HCL2 over to it.
func getCoreBuildProvisioners(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, blocks []*hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		if pb.OnlyExcept.Skip(source.String()) {
			continue
		}

		coreBuildProv, moreDiags := getCoreBuildProvisioner(cfg, source, pb, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, coreBuildProv)
	}
	return res, diags
}

func startProvisioner(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, pb *hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) (packersdk.Provisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	provisioner, err := cfg.Parser.PluginConfig.Provisioners.Start(pb.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed loading %s", pb.PType),
			Subject:  pb.HCL2Ref.LabelsRanges[0].Ptr(),
			Detail:   err.Error(),
		})
		return nil, diags
	}

	builderVars := source.BuilderVariables()
	builderVars["packer_core_version"] = cfg.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(cfg.Debug)
	builderVars["packer_force"] = strconv.FormatBool(cfg.Force)
	builderVars["packer_on_error"] = cfg.OnError

	hclProvisioner := &hcl2template.HCL2Provisioner{
		Provisioner:      provisioner,
		ProvisionerBlock: pb,
		EvalContext:      ectx,
		BuilderVariables: builderVars,
	}

	if pb.Override != nil {
		if override, ok := pb.Override[source.Name()]; ok {
			hclProvisioner.Override = override.(map[string]interface{})
		}
	}

	err = hclProvisioner.HCL2Prepare(nil)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed preparing %s", pb),
			Detail:   err.Error(),
			Subject:  pb.HCL2Ref.DefRange.Ptr(),
		})
		return nil, diags
	}
	return hclProvisioner, diags
}

func getCoreBuildProvisioner(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, pb *hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	provisioner, moreDiags := startProvisioner(cfg, source, pb, ectx)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return packer.CoreBuildProvisioner{}, diags
	}

	log.Printf("[PROVISIONER] original HCL2 body: %#v", pb.HCL2Ref.Rest)
	flatProvisionerCfg, diags := hcl2template.DecodeHCL2Spec(pb.HCL2Ref.Rest, ectx, provisioner)
	log.Printf("[PROVISIONER] flattened HCL2 config: %#v", flatProvisionerCfg)
	if diags.HasErrors() {
		return packer.CoreBuildProvisioner{}, diags
	}

	// If we're pausing, we wrap the provisioner in a special pauser.
	if pb.PauseBefore != 0 {
		provisioner = &packer.PausedProvisioner{
			PauseBefore: pb.PauseBefore,
			Provisioner: provisioner,
		}
	} else if pb.Timeout != 0 {
		provisioner = &packer.TimeoutProvisioner{
			Timeout:     pb.Timeout,
			Provisioner: provisioner,
		}
	}
	if pb.MaxRetries != 0 {
		provisioner = &packer.RetriedProvisioner{
			MaxRetries:  pb.MaxRetries,
			Provisioner: provisioner,
		}
	}

	return packer.CoreBuildProvisioner{
		PType:       pb.PType,
		PName:       pb.PName,
		Provisioner: provisioner,
		HCLConfig:   flatProvisionerCfg,
	}, diags
}

func startPostProcessor(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, pp *hcl2template.PostProcessorBlock, ectx *hcl.EvalContext) (packersdk.PostProcessor, hcl.Diagnostics) {
	// ProvisionerBlock represents a detected but unparsed provisioner
	var diags hcl.Diagnostics

	postProcessor, err := cfg.Parser.PluginConfig.PostProcessors.Start(pp.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed loading %s", pp.PType),
			Subject:  pp.DefRange.Ptr(),
			Detail:   err.Error(),
		})
		return nil, diags
	}

	builderVars := source.BuilderVariables()
	builderVars["packer_core_version"] = cfg.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(cfg.Debug)
	builderVars["packer_force"] = strconv.FormatBool(cfg.Force)
	builderVars["packer_on_error"] = cfg.OnError

	hclPostProcessor := &hcl2template.HCL2PostProcessor{
		PostProcessor:      postProcessor,
		PostProcessorBlock: pp,
		EvalContext:        ectx,
		BuilderVariables:   builderVars,
	}
	err = hclPostProcessor.HCL2Prepare(nil)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed preparing %s", pp),
			Detail:   err.Error(),
			Subject:  pp.DefRange.Ptr(),
		})
		return nil, diags
	}
	return hclPostProcessor, diags
}

// getCoreBuildProvisioners takes a list of post processor block, starts
// according provisioners and sends parsed HCL2 over to it.
func getCoreBuildPostProcessors(cfg *hcl2template.PackerConfig, source hcl2template.SourceUseBlock, blocksList [][]*hcl2template.PostProcessorBlock, ectx *hcl.EvalContext, exceptMatches *int) ([][]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := [][]packer.CoreBuildPostProcessor{}
	for _, blocks := range blocksList {
		pps := []packer.CoreBuildPostProcessor{}
		for _, ppb := range blocks {
			if ppb.OnlyExcept.Skip(source.String()) {
				continue
			}

			name := ppb.PName
			if name == "" {
				name = ppb.PType
			}
			// -except
			exclude := false
			for _, exceptGlob := range cfg.Except {
				if exceptGlob.Match(name) {
					exclude = true
					*exceptMatches = *exceptMatches + 1
					break
				}
			}
			if exclude {
				break
			}

			postProcessor, moreDiags := startPostProcessor(cfg, source, ppb, ectx)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			flatPostProcessorCfg, moreDiags := hcl2template.DecodeHCL2Spec(ppb.HCL2Ref.Rest, ectx, postProcessor)

			pps = append(pps, packer.CoreBuildPostProcessor{
				PostProcessor:     postProcessor,
				PName:             ppb.PName,
				PType:             ppb.PType,
				HCLConfig:         flatPostProcessorCfg,
				KeepInputArtifact: ppb.KeepInputArtifact,
			})
		}
		if len(pps) > 0 {
			res = append(res, pps)
		}
	}

	return res, diags
}

// GetBuilds returns a list of packer Build based on the HCL2 parsed build
// blocks. All Builders, Provisioners and Post Processors will be started and
// configured.
func (s *SequentialScheduler) hcl2GetBuilds(opts packer.GetBuildsOptions) ([]packersdk.Build, hcl.Diagnostics) {
	cfg := s.handler.(*hcl2template.PackerConfig)

	res := []packersdk.Build{}
	var diags hcl.Diagnostics
	possibleBuildNames := []string{}

	cfg.Debug = opts.Debug
	cfg.Force = opts.Force
	cfg.OnError = opts.OnError

	if len(cfg.Builds) == 0 {
		return res, append(diags, &hcl.Diagnostic{
			Summary:  "Missing build block",
			Detail:   "A build block with one or more sources is required for executing a build.",
			Severity: hcl.DiagError,
		})
	}

	for _, build := range cfg.Builds {
		for _, srcUsage := range build.Sources {
			src, found := cfg.Sources[srcUsage.SourceRef]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + hcl2template.SourceLabel + " " + srcUsage.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   fmt.Sprintf("Known: %v", cfg.Sources),
				})
				continue
			}

			pcb := &packer.CoreBuild{
				BuildName: build.Name,
				Type:      srcUsage.String(),
			}

			pcb.SetDebug(cfg.Debug)
			pcb.SetForce(cfg.Force)
			pcb.SetOnError(cfg.OnError)

			// Apply the -only and -except command-line options to exclude matching builds.
			buildName := pcb.Name()
			possibleBuildNames = append(possibleBuildNames, buildName)
			// -only
			if len(opts.Only) > 0 {
				onlyGlobs, diags := convertFilterOption(opts.Only, "only")
				if diags.HasErrors() {
					return nil, diags
				}
				cfg.Only = onlyGlobs
				include := false
				for _, onlyGlob := range onlyGlobs {
					if onlyGlob.Match(buildName) {
						include = true
						break
					}
				}
				if !include {
					continue
				}
				opts.OnlyMatches++
			}

			// -except
			if len(opts.Except) > 0 {
				exceptGlobs, diags := convertFilterOption(opts.Except, "except")
				if diags.HasErrors() {
					return nil, diags
				}
				cfg.Except = exceptGlobs
				exclude := false
				for _, exceptGlob := range exceptGlobs {
					if exceptGlob.Match(buildName) {
						exclude = true
						break
					}
				}
				if exclude {
					opts.ExceptMatches++
					continue
				}
			}

			builder, moreDiags, generatedVars := startBuilder(cfg, srcUsage, cfg.EvalContext(hcl2template.BuildContext, nil))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			decoded, _ := hcl2template.DecodeHCL2Spec(srcUsage.Body, cfg.EvalContext(hcl2template.BuildContext, nil), builder)
			pcb.HCLConfig = decoded

			// If the builder has provided a list of to-be-generated variables that
			// should be made accessible to provisioners, pass that list into
			// the provisioner prepare() so that the provisioner can appropriately
			// validate user input against what will become available. Otherwise,
			// only pass the default variables, using the basic placeholder data.
			unknownBuildValues := map[string]cty.Value{}
			for _, k := range append(packer.BuilderDataCommonKeys, generatedVars...) {
				unknownBuildValues[k] = cty.StringVal("<unknown>")
			}
			unknownBuildValues["name"] = cty.StringVal(build.Name)

			variables := map[string]cty.Value{
				hcl2template.SourcesAccessor: cty.ObjectVal(srcUsage.CtyValues()),
				hcl2template.BuildAccessor:   cty.ObjectVal(unknownBuildValues),
			}

			provisioners, moreDiags := getCoreBuildProvisioners(cfg, srcUsage, build.ProvisionerBlocks, cfg.EvalContext(hcl2template.BuildContext, variables))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			pps, moreDiags := getCoreBuildPostProcessors(cfg, srcUsage, build.PostProcessorsLists, cfg.EvalContext(hcl2template.BuildContext, variables), &opts.ExceptMatches)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			if build.ErrorCleanupProvisionerBlock != nil &&
				!build.ErrorCleanupProvisionerBlock.OnlyExcept.Skip(srcUsage.String()) {
				errorCleanupProv, moreDiags := getCoreBuildProvisioner(cfg, srcUsage, build.ErrorCleanupProvisionerBlock, cfg.EvalContext(hcl2template.BuildContext, variables))
				diags = append(diags, moreDiags...)
				if moreDiags.HasErrors() {
					continue
				}
				pcb.CleanupProvisioner = errorCleanupProv
			}

			pcb.Builder = builder
			pcb.Provisioners = provisioners
			pcb.PostProcessors = pps
			pcb.Prepared = true

			// Prepare just sets the "prepareCalled" flag on CoreBuild, since
			// we did all the prep here.
			_, err := pcb.Prepare()
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Preparing packer core build %s failed", src.Ref().String()),
					Detail:   err.Error(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
				})
				continue
			}

			res = append(res, pcb)
		}
	}
	if len(opts.Only) > opts.OnlyMatches {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "an 'only' option was passed, but not all matches were found for the given build.",
			Detail: fmt.Sprintf("Possible build names: %v.\n"+
				"These could also be matched with a glob pattern like: 'happycloud.*'", possibleBuildNames),
		})
	}
	if len(opts.Except) > opts.ExceptMatches {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "an 'except' option was passed, but did not match any build.",
			Detail: fmt.Sprintf("Possible build names: %v.\n"+
				"These could also be matched with a glob pattern like: 'happycloud.*'", possibleBuildNames),
		})
	}
	return res, diags
}

// BuildNames returns the builds that are available in this configured core.
func BuildNames(c *packer.Core, only, except []string) []string {

	sort.Strings(only)
	sort.Strings(except)
	c.Except = except
	c.Only = only

	r := make([]string, 0, len(c.Builds))
	for n := range c.Builds {
		onlyPos := sort.SearchStrings(only, n)
		foundInOnly := onlyPos < len(only) && only[onlyPos] == n
		if len(only) > 0 && !foundInOnly {
			continue
		}

		if pos := sort.SearchStrings(except, n); pos < len(except) && except[pos] == n {
			continue
		}
		r = append(r, n)
	}
	sort.Strings(r)

	return r
}

func generateCoreBuildProvisioner(c *packer.Core, rawP *template.Provisioner, rawName string) (packer.CoreBuildProvisioner, error) {
	// Get the provisioner
	cbp := packer.CoreBuildProvisioner{}
	provisioner, err := c.Components.PluginConfig.Provisioners.Start(rawP.Type)
	if err != nil {
		return cbp, fmt.Errorf(
			"error initializing provisioner '%s': %s",
			rawP.Type, err)
	}
	if provisioner == nil {
		return cbp, fmt.Errorf(
			"provisioner type not found: %s", rawP.Type)
	}

	// Get the configuration
	config := make([]interface{}, 1, 2)
	config[0] = rawP.Config
	if rawP.Override != nil {
		if override, ok := rawP.Override[rawName]; ok {
			config = append(config, override)
		}
	}
	// If we're pausing, we wrap the provisioner in a special pauser.
	if rawP.PauseBefore != 0 {
		provisioner = &packer.PausedProvisioner{
			PauseBefore: rawP.PauseBefore,
			Provisioner: provisioner,
		}
	} else if rawP.Timeout != 0 {
		provisioner = &packer.TimeoutProvisioner{
			Timeout:     rawP.Timeout,
			Provisioner: provisioner,
		}
	}
	maxRetries := 0
	if rawP.MaxRetries != "" {
		renderedMaxRetries, err := interpolate.Render(rawP.MaxRetries, c.Context())
		if err != nil {
			return cbp, fmt.Errorf("failed to interpolate `max_retries`: %s", err.Error())
		}
		maxRetries, err = strconv.Atoi(renderedMaxRetries)
		if err != nil {
			return cbp, fmt.Errorf("`max_retries` must be a valid integer: %s", err.Error())
		}
	}
	if maxRetries != 0 {
		provisioner = &packer.RetriedProvisioner{
			MaxRetries:  maxRetries,
			Provisioner: provisioner,
		}
	}
	cbp = packer.CoreBuildProvisioner{
		PType:       rawP.Type,
		Provisioner: provisioner,
		Config:      config,
	}

	return cbp, nil
}

// This is used for json templates to launch the build plugins.
// They will be prepared via b.Prepare() later.
func (s *SequentialScheduler) jsonGetBuilds(opts packer.GetBuildsOptions) ([]packersdk.Build, hcl.Diagnostics) {
	c := s.handler.(*packer.Core)

	buildNames := BuildNames(c, opts.Only, opts.Except)
	builds := []packersdk.Build{}
	diags := hcl.Diagnostics{}
	for _, n := range buildNames {
		b, err := Build(c, n)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to initialize build %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Now that build plugin has been launched, call Prepare()
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(opts.Debug)
		b.SetForce(opts.Force)
		b.SetOnError(opts.OnError)

		warnings, err := b.Prepare()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to prepare build: %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Only append builds to list if the Prepare() is successful.
		builds = append(builds, b)

		if len(warnings) > 0 {
			for _, warning := range warnings {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf("Warning when preparing build: %q", n),
					Detail:   warning,
				})
			}
		}
	}
	return builds, diags
}

// Build returns the Build object for the given name.
func Build(c *packer.Core, n string) (packersdk.Build, error) {
	// Setup the builder
	configBuilder, ok := c.Builds[n]
	if !ok {
		return nil, fmt.Errorf("no such build found: %s", n)
	}
	// BuilderStore = config.Builders, gathered in loadConfig() in main.go
	// For reference, the builtin BuilderStore is generated in
	// packer/config.go in the Discover() func.

	// the Start command launches the builder plugin of the given type without
	// calling Prepare() or passing any build-specific details.
	builder, err := c.Components.PluginConfig.Builders.Start(configBuilder.Type)
	if err != nil {
		return nil, fmt.Errorf(
			"error initializing builder '%s': %s",
			configBuilder.Type, err)
	}
	if builder == nil {
		return nil, fmt.Errorf(
			"builder type not found: %s", configBuilder.Type)
	}

	// rawName is the uninterpolated name that we use for various lookups
	rawName := configBuilder.Name

	// Setup the provisioners for this build
	provisioners := make([]packer.CoreBuildProvisioner, 0, len(c.Template.Provisioners))
	for _, rawP := range c.Template.Provisioners {
		// If we're skipping this, then ignore it
		if rawP.OnlyExcept.Skip(rawName) {
			continue
		}
		cbp, err := generateCoreBuildProvisioner(c, rawP, rawName)
		if err != nil {
			return nil, err
		}

		provisioners = append(provisioners, cbp)
	}

	var cleanupProvisioner packer.CoreBuildProvisioner
	if c.Template.CleanupProvisioner != nil {
		// This is a special instantiation of the shell-local provisioner that
		// is only run on error at end of provisioning step before other step
		// cleanup occurs.
		cleanupProvisioner, err = generateCoreBuildProvisioner(c, c.Template.CleanupProvisioner, rawName)
		if err != nil {
			return nil, err
		}
	}

	// Setup the post-processors
	postProcessors := make([][]packer.CoreBuildPostProcessor, 0, len(c.Template.PostProcessors))
	for _, rawPs := range c.Template.PostProcessors {
		current := make([]packer.CoreBuildPostProcessor, 0, len(rawPs))
		for _, rawP := range rawPs {
			if rawP.Skip(rawName) {
				continue
			}
			// -except skips post-processor & build
			foundExcept := false
			for _, except := range c.Except {
				if except != "" && except == rawP.Name {
					foundExcept = true
				}
			}
			if foundExcept {
				break
			}

			// Get the post-processor
			postProcessor, err := c.Components.PluginConfig.PostProcessors.Start(rawP.Type)
			if err != nil {
				return nil, fmt.Errorf(
					"error initializing post-processor '%s': %s",
					rawP.Type, err)
			}
			if postProcessor == nil {
				return nil, fmt.Errorf(
					"post-processor type not found: %s", rawP.Type)
			}

			current = append(current, packer.CoreBuildPostProcessor{
				PostProcessor:     postProcessor,
				PType:             rawP.Type,
				PName:             rawP.Name,
				Config:            rawP.Config,
				KeepInputArtifact: rawP.KeepInputArtifact,
			})
		}

		// If we have no post-processors in this chain, just continue.
		if len(current) == 0 {
			continue
		}

		postProcessors = append(postProcessors, current)
	}

	// Return a structure that contains the plugins, their types, variables, and
	// the raw builder config loaded from the json template
	cb := &packer.CoreBuild{
		Type:               n,
		Builder:            builder,
		BuilderConfig:      configBuilder.Config,
		BuilderType:        configBuilder.Type,
		PostProcessors:     postProcessors,
		Provisioners:       provisioners,
		CleanupProvisioner: cleanupProvisioner,
		TemplatePath:       c.Template.Path,
		Variables:          c.Variables,
	}

	//configBuilder.Name is left uninterpolated so we must check against
	// the interpolated name.
	if configBuilder.Type != configBuilder.Name {
		cb.BuildName = configBuilder.Type
	}

	return cb, nil
}

func filterVarsFromLogs(inputOrLocal hcl2template.Variables) {
	for _, variable := range inputOrLocal {
		if !variable.Sensitive {
			continue
		}
		value := variable.Value()
		_ = cty.Walk(value, func(_ cty.Path, nested cty.Value) (bool, error) {
			if nested.IsWhollyKnown() && !nested.IsNull() && nested.Type().Equals(cty.String) {
				packersdk.LogSecretFilter.Set(nested.AsString())
			}
			return true, nil
		})
	}
}

func (s *SequentialScheduler) hcl2EvaluateLocalVariables() hcl.Diagnostics {
	c := s.handler.(*hcl2template.PackerConfig)
	locals := c.LocalBlocks

	var diags hcl.Diagnostics

	if len(locals) == 0 {
		return diags
	}

	if c.LocalVariables == nil {
		c.LocalVariables = hcl2template.Variables{}
	}

	for foundSomething := true; foundSomething; {
		foundSomething = false
		for i := 0; i < len(locals); {
			local := locals[i]
			moreDiags := hcl2EvaluateLocalVariable(c, local)
			if moreDiags.HasErrors() {
				i++
				continue
			}
			foundSomething = true
			locals = append(locals[:i], locals[i+1:]...)
		}
	}

	if len(locals) != 0 {
		// get errors from remaining variables
		return hcl2EvaluateAllLocalVariables(c, locals)
	}

	filterVarsFromLogs(c.InputVariables)
	filterVarsFromLogs(c.LocalVariables)

	return diags
}

func hcl2EvaluateAllLocalVariables(c *hcl2template.PackerConfig, locals []*hcl2template.LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, local := range locals {
		diags = append(diags, hcl2EvaluateLocalVariable(c, local)...)
	}

	return diags
}

func hcl2EvaluateLocalVariable(c *hcl2template.PackerConfig, local *hcl2template.LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	value, moreDiags := local.Expr.Value(c.EvalContext(hcl2template.LocalContext, nil))
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return diags
	}
	c.LocalVariables[local.Name] = &hcl2template.Variable{
		Name:      local.Name,
		Sensitive: local.Sensitive,
		Values: []hcl2template.VariableAssignment{{
			Value: value,
			Expr:  local.Expr,
			From:  "default",
		}},
		Type: value.Type(),
	}

	return diags
}

func (s *SequentialScheduler) jsonVariableEval() hcl.Diagnostics {
	core := s.handler.(*packer.Core)
	if err := initJSON(core); err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to evaluate variables",
				Detail:   err.Error(),
			},
		}
	}
	for _, secret := range core.Secrets {
		packersdk.LogSecretFilter.Set(secret)
	}

	var diags hcl.Diagnostics

	// Go through and interpolate all the build names. We should be able
	// to do this at this point with the variables.
	core.Builds = make(map[string]*template.Builder)
	for _, b := range core.Template.Builders {
		v, err := interpolate.Render(b.Name, core.Context())
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Build interpolation failure",
				Detail: fmt.Sprintf("Error interpolating builder '%s': %s",
					b.Name, err),
			})
		}

		core.Builds[v] = b
	}

	return diags
}

func initJSON(c *packer.Core) error {
	if c.Variables == nil {
		c.Variables = make(map[string]string)
	}
	// Go through the variables and interpolate the environment and
	// user variables
	ctx, err := renderVarsRecursively(c)
	if err != nil {
		return err
	}
	for _, v := range c.Template.SensitiveVariables {
		secret := ctx.UserVariables[v.Key]
		c.Secrets = append(c.Secrets, secret)
	}

	return nil
}

func renderVarsRecursively(c *packer.Core) (*interpolate.Context, error) {
	ctx := c.Context()
	ctx.EnableEnv = true
	ctx.UserVariables = make(map[string]string)
	shouldRetry := true
	changed := false
	failedInterpolation := ""

	// Why this giant loop?  User variables can be recursively defined. For
	// example:
	// "variables": {
	//    	"foo":  "bar",
	//	 	"baz":  "{{user `foo`}}baz",
	// 		"bang": "bang{{user `baz`}}"
	// },
	// In this situation, we cannot guarantee that we've added "foo" to
	// UserVariables before we try to interpolate "baz" the first time. We need
	// to have the option to loop back over in order to add the properly
	// interpolated "baz" to the UserVariables map.
	// Likewise, we'd need to loop up to two times to properly add "bang",
	// since that depends on "baz" being set, which depends on "foo" being set.

	// We break out of the while loop either if all our variables have been
	// interpolated or if after 100 loops we still haven't succeeded in
	// interpolating them.  Please don't actually nest your variables in 100
	// layers of other variables. Please.

	// c.Template.Variables is populated by variables defined within the Template
	// itself
	// c.variables is populated by variables read in from the command line and
	// var-files.
	// We need to read the keys from both, then loop over all of them to figure
	// out the appropriate interpolations.

	repeatMap := make(map[string]string)
	allKeys := make([]string, 0)

	// load in template variables
	for k, v := range c.Template.Variables {
		repeatMap[k] = v.Default
		allKeys = append(allKeys, k)
	}

	// overwrite template variables with command-line-read variables
	for k, v := range c.Variables {
		repeatMap[k] = v
		allKeys = append(allKeys, k)
	}

	// sort map to force the following loop to be deterministic.
	sort.Strings(allKeys)
	type keyValue struct {
		Key   string
		Value string
	}
	sortedMap := make([]keyValue, len(repeatMap))
	for _, k := range allKeys {
		sortedMap = append(sortedMap, keyValue{k, repeatMap[k]})
	}

	// Regex to exclude any build function variable or template variable
	// from interpolating earlier
	// E.g.: {{ .HTTPIP }}  won't interpolate now
	renderFilter := "{{(\\s|)\\.(.*?)(\\s|)}}"

	for i := 0; i < 100; i++ {
		shouldRetry = false
		changed = false
		deleteKeys := []string{}
		// First, loop over the variables in the template
		for _, kv := range sortedMap {
			// Interpolate the default
			renderedV, err := interpolate.RenderRegex(kv.Value, ctx, renderFilter)
			switch err.(type) {
			case nil:
				// We only get here if interpolation has succeeded, so something is
				// different in this loop than in the last one.
				changed = true
				c.Variables[kv.Key] = renderedV
				ctx.UserVariables = c.Variables
				// Remove fully-interpolated variables from the map, and flag
				// variables that still need interpolating for a repeat.
				done, err := isDoneInterpolating(kv.Value)
				if err != nil {
					return ctx, err
				}
				if done {
					deleteKeys = append(deleteKeys, kv.Key)
				} else {
					shouldRetry = true
				}
			case ttmp.ExecError:
				castError := err.(ttmp.ExecError)
				if strings.Contains(castError.Error(), interpolate.ErrVariableNotSetString) {
					shouldRetry = true
					failedInterpolation = fmt.Sprintf(`"%s": "%s"; error: %s`, kv.Key, kv.Value, err)
				} else {
					return ctx, err
				}
			default:
				return ctx, fmt.Errorf(
					// unexpected interpolation error: abort the run
					"error interpolating default value for '%s': %s",
					kv.Key, err)
			}
		}
		if !shouldRetry {
			break
		}

		// Clear completed vars from sortedMap before next loop. Do this one
		// key at a time because the indices are gonna change ever time you
		// delete from the map.
		for _, k := range deleteKeys {
			for ind, kv := range sortedMap {
				if kv.Key == k {
					sortedMap = append(sortedMap[:ind], sortedMap[ind+1:]...)
					break
				}
			}
		}
		deleteKeys = []string{}
	}

	if !changed && shouldRetry {
		return ctx, fmt.Errorf("Failed to interpolate %s: Please make sure that "+
			"the variable you're referencing has been defined; Packer treats "+
			"all variables used to interpolate other user variables as "+
			"required.", failedInterpolation)
	}

	return ctx, nil
}

func isDoneInterpolating(v string) (bool, error) {
	// Check for whether the var contains any more references to `user`, wrapped
	// in interpolation syntax.
	filter := `{{\s*user\s*\x60.*\x60\s*}}`
	matched, err := regexp.MatchString(filter, v)
	if err != nil {
		return false, fmt.Errorf("Can't tell if interpolation is done: %s", err)
	}
	if matched {
		// not done interpolating; there's still a call to "user" in a template
		// engine
		return false, nil
	}
	// No more calls to "user" as a template engine, so we're done.
	return true, nil
}

func (s *SequentialScheduler) EvaluateDataSources() hcl.Diagnostics {
	switch cfg := s.handler.(type) {
	// Legacy JSON templates do not have datasources
	case *packer.Core:
		return nil
	case *hcl2template.PackerConfig:
		return evaluateDatasources(cfg, s.skipDatasourcesExecution)
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown configuration type",
			Detail: `
The packer handler is of unknown type %q, expected either a *packer.Core or a *hcl2template.PackerConfig

This is likely a Packer bug, please report this so the team can take a look at it.`,
		},
	}
}

func evaluateDatasources(cfg *hcl2template.PackerConfig, skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	dependencies := map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef{}
	for ref, ds := range cfg.Datasources {
		if ds.Value != (cty.Value{}) {
			continue
		}
		// Pre-examine body of this data source to see if it uses another data
		// source in any of its input expressions. If so, skip evaluating it for
		// now, and add it to a list of datasources to evaluate again, later,
		// with the datasources in its context.
		// This is essentially creating a very primitive DAG just for data
		// source interdependencies.
		block := ds.Block
		body := block.Body
		attrs, _ := body.JustAttributes()

		skipFirstEval := false
		for _, attr := range attrs {
			vars := attr.Expr.Variables()
			for _, v := range vars {
				// check whether the variable is a data source
				if v.RootName() == "data" {
					// construct, backwards, the data source type and name we
					// need to evaluate before this one can be evaluated.
					dependsOn := hcl2template.DatasourceRef{
						Type: v[1].(hcl.TraverseAttr).Name,
						Name: v[2].(hcl.TraverseAttr).Name,
					}
					log.Printf("The data source %#v depends on datasource %#v", ref, dependsOn)
					if dependencies[ref] != nil {
						dependencies[ref] = append(dependencies[ref], dependsOn)
					} else {
						dependencies[ref] = []hcl2template.DatasourceRef{dependsOn}
					}
					skipFirstEval = true
				}
			}
		}

		// Now we have a list of data sources that depend on other data sources.
		// Don't evaluate these; only evaluate data sources that we didn't
		// mark  as having dependencies.
		if skipFirstEval {
			continue
		}

		datasource, startDiags := cfg.StartDatasource(cfg.Parser.PluginConfig.DataSources, ref, false)
		diags = append(diags, startDiags...)
		if diags.HasErrors() {
			continue
		}

		if skipExecution {
			placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
			ds.Value = placeholderValue
			cfg.Datasources[ref] = ds
			continue
		}

		dsOpts, _ := hcl2template.DecodeHCL2Spec(body, cfg.EvalContext(hcl2template.DatasourceContext, nil), datasource)
		sp := packer.CheckpointReporter.AddSpan(ref.Type, "datasource", dsOpts)
		realValue, err := datasource.Execute()
		sp.End(err)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  err.Error(),
				Subject:  &cfg.Datasources[ref].Block.DefRange,
				Severity: hcl.DiagError,
			})
			continue
		}

		ds.Value = realValue
		cfg.Datasources[ref] = ds
	}

	// Now that most of our data sources have been started and executed, we can
	// try to execute the ones that depend on other data sources.
	for ref := range dependencies {
		_, moreDiags, _ := recursivelyEvaluateDatasources(cfg, ref, dependencies, skipExecution, 0)
		// Deduplicate diagnostics to prevent recursion messes.
		cleanedDiags := map[string]*hcl.Diagnostic{}
		for _, diag := range moreDiags {
			cleanedDiags[diag.Summary] = diag
		}

		for _, diag := range cleanedDiags {
			diags = append(diags, diag)
		}
	}

	return diags
}

func recursivelyEvaluateDatasources(
	cfg *hcl2template.PackerConfig,
	ref hcl2template.DatasourceRef,
	dependencies map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef,
	skipExecution bool,
	depth int,
) (map[hcl2template.DatasourceRef][]hcl2template.DatasourceRef, hcl.Diagnostics, bool) {
	var diags hcl.Diagnostics
	var moreDiags hcl.Diagnostics
	shouldContinue := true

	if depth > 10 {
		// Add a comment about recursion.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Max datasource recursion depth exceeded.",
			Detail: "An error occured while recursively evaluating data " +
				"sources. Either your data source depends on more than ten " +
				"other data sources, or your data sources have a cyclic " +
				"dependency. Please simplify your config to continue. ",
		})
		return dependencies, diags, false
	}

	ds := cfg.Datasources[ref]
	// Make sure everything ref depends on has already been evaluated.
	for _, dep := range dependencies[ref] {
		if _, ok := dependencies[dep]; ok {
			depth += 1
			// If this dependency is not in the map, it means we've already
			// launched and executed this datasource. Otherwise, it means
			// we still need to run it. RECURSION TIME!!
			dependencies, moreDiags, shouldContinue = recursivelyEvaluateDatasources(cfg, dep, dependencies, skipExecution, depth)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				return dependencies, diags, shouldContinue
			}
		}
	}
	// If we've gotten here, then it means ref doesn't seem to have any further
	// dependencies we need to evaluate first. Evaluate it, with the cfg's full
	// data source context.
	datasource, startDiags := cfg.StartDatasource(cfg.Parser.PluginConfig.DataSources, ref, true)
	if startDiags.HasErrors() {
		diags = append(diags, startDiags...)
		return dependencies, diags, shouldContinue
	}

	if skipExecution {
		placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
		ds.Value = placeholderValue
		cfg.Datasources[ref] = ds
		return dependencies, diags, shouldContinue
	}

	opts, _ := hcl2template.DecodeHCL2Spec(ds.Block.Body, cfg.EvalContext(hcl2template.DatasourceContext, nil), datasource)
	sp := packer.CheckpointReporter.AddSpan(ref.Type, "datasource", opts)
	realValue, err := datasource.Execute()
	sp.End(err)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &cfg.Datasources[ref].Block.DefRange,
			Severity: hcl.DiagError,
		})
		return dependencies, diags, shouldContinue
	}

	ds.Value = realValue
	cfg.Datasources[ref] = ds
	// remove ref from the dependencies map.
	delete(dependencies, ref)
	return dependencies, diags, shouldContinue
}
