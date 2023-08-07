package schedulers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gobwas/glob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer-plugin-sdk/didyoumean"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func (s *HCLSequentialScheduler) EvaluateBuilds() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// parse the actual content // rest
	for _, file := range s.config.Files {
		diags = append(diags, s.config.Parser.ParseConfig(file, s.config)...)
	}

	diags = append(diags, s.initializeBlocks()...)

	return diags
}

func (s *HCLSequentialScheduler) initializeBlocks() hcl.Diagnostics {
	// verify that all used plugins do exist
	var diags hcl.Diagnostics

	for _, build := range s.config.Builds {
		for i := range build.Sources {
			// here we grab a pointer to the source usage because we will set
			// its body.
			srcUsage := &(build.Sources[i])
			if !s.config.Parser.PluginConfig.Builders.Has(srcUsage.Type) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + hcl2template.BuildSourceLabel + " type " + srcUsage.Type,
					Subject:  &build.HCL2Ref.DefRange,
					Detail:   fmt.Sprintf("known builders: %v", s.config.Parser.PluginConfig.Builders.List()),
					Severity: hcl.DiagError,
				})
				continue
			}

			sourceDefinition, found := s.config.Sources[srcUsage.SourceRef]
			if !found {
				availableSrcs := hcl2template.ListAvailableSourceNames(s.config.Sources)
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
			if !s.config.Parser.PluginConfig.Provisioners.Has(provBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildProvisionerLabel+" type %q", provBlock.PType),
					Subject:  provBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+hcl2template.BuildProvisionerLabel+"s: %v", s.config.Parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
		}

		if build.ErrorCleanupProvisionerBlock != nil {
			if !s.config.Parser.PluginConfig.Provisioners.Has(build.ErrorCleanupProvisionerBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildErrorCleanupProvisionerLabel+" type %q", build.ErrorCleanupProvisionerBlock.PType),
					Subject:  build.ErrorCleanupProvisionerBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+hcl2template.BuildErrorCleanupProvisionerLabel+"s: %v", s.config.Parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
		}

		for _, ppList := range build.PostProcessorsLists {
			for _, ppBlock := range ppList {
				if !s.config.Parser.PluginConfig.PostProcessors.Has(ppBlock.PType) {
					diags = append(diags, &hcl.Diagnostic{
						Summary:  fmt.Sprintf("Unknown "+hcl2template.BuildPostProcessorLabel+" type %q", ppBlock.PType),
						Subject:  ppBlock.HCL2Ref.TypeRange.Ptr(),
						Detail:   fmt.Sprintf("known "+hcl2template.BuildPostProcessorLabel+"s: %v", s.config.Parser.PluginConfig.PostProcessors.List()),
						Severity: hcl.DiagError,
					})
				}
			}
		}

	}

	return diags
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

func (s *HCLSequentialScheduler) startBuilder(source hcl2template.SourceUseBlock, ectx *hcl.EvalContext) (packersdk.Builder, hcl.Diagnostics, []string) {
	var diags hcl.Diagnostics

	builder, err := s.config.Parser.PluginConfig.Builders.Start(source.Type)
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
	builderVars["packer_core_version"] = s.config.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(s.opts.Debug)
	builderVars["packer_force"] = strconv.FormatBool(s.opts.Force)
	builderVars["packer_on_error"] = s.opts.OnError

	generatedVars, warning, err := builder.Prepare(builderVars, decoded)
	moreDiags = warningErrorsToDiags(s.config.Sources[source.SourceRef].Block, warning, err)
	diags = append(diags, moreDiags...)
	return builder, diags, generatedVars
}

// getCoreBuildProvisioners takes a list of provisioner block, starts according
// provisioners and sends parsed HCL2 over to it.
func (s *HCLSequentialScheduler) getCoreBuildProvisioners(source hcl2template.SourceUseBlock, blocks []*hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		if pb.OnlyExcept.Skip(source.String()) {
			continue
		}

		coreBuildProv, moreDiags := s.getCoreBuildProvisioner(source, pb, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, coreBuildProv)
	}
	return res, diags
}

func (s *HCLSequentialScheduler) startProvisioner(source hcl2template.SourceUseBlock, pb *hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) (packersdk.Provisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	provisioner, err := s.config.Parser.PluginConfig.Provisioners.Start(pb.PType)
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
	builderVars["packer_core_version"] = s.config.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(s.config.Debug)
	builderVars["packer_force"] = strconv.FormatBool(s.config.Force)
	builderVars["packer_on_error"] = s.config.OnError

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

func (s *HCLSequentialScheduler) getCoreBuildProvisioner(source hcl2template.SourceUseBlock, pb *hcl2template.ProvisionerBlock, ectx *hcl.EvalContext) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	provisioner, moreDiags := s.startProvisioner(source, pb, ectx)
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

func (s *HCLSequentialScheduler) startPostProcessor(source hcl2template.SourceUseBlock, pp *hcl2template.PostProcessorBlock, ectx *hcl.EvalContext) (packersdk.PostProcessor, hcl.Diagnostics) {
	// ProvisionerBlock represents a detected but unparsed provisioner
	var diags hcl.Diagnostics

	postProcessor, err := s.config.Parser.PluginConfig.PostProcessors.Start(pp.PType)
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
	builderVars["packer_core_version"] = s.config.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(s.opts.Debug)
	builderVars["packer_force"] = strconv.FormatBool(s.opts.Force)
	builderVars["packer_on_error"] = s.opts.OnError

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
func (s *HCLSequentialScheduler) getCoreBuildPostProcessors(source hcl2template.SourceUseBlock, blocksList [][]*hcl2template.PostProcessorBlock, ectx *hcl.EvalContext, exceptMatches *int) ([][]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
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
			for _, exceptGlob := range s.config.Except {
				if exceptGlob.Match(name) {
					exclude = true
					*exceptMatches = *exceptMatches + 1
					break
				}
			}
			if exclude {
				break
			}

			postProcessor, moreDiags := s.startPostProcessor(source, ppb, ectx)
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
func (s *HCLSequentialScheduler) GetBuilds() ([]packersdk.Build, hcl.Diagnostics) {
	res := []packersdk.Build{}
	var diags hcl.Diagnostics
	possibleBuildNames := []string{}

	if len(s.config.Builds) == 0 {
		return res, append(diags, &hcl.Diagnostic{
			Summary:  "Missing build block",
			Detail:   "A build block with one or more sources is required for executing a build.",
			Severity: hcl.DiagError,
		})
	}

	onlyMatches := 0
	exceptMatches := 0

	for _, build := range s.config.Builds {
		for _, srcUsage := range build.Sources {
			src, found := s.config.Sources[srcUsage.SourceRef]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + hcl2template.SourceLabel + " " + srcUsage.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   fmt.Sprintf("Known: %v", s.config.Sources),
				})
				continue
			}

			pcb := &packer.CoreBuild{
				BuildName: build.Name,
				Type:      srcUsage.String(),
			}

			pcb.SetDebug(s.opts.Debug)
			pcb.SetForce(s.opts.Force)
			pcb.SetOnError(s.opts.OnError)

			// Apply the -only and -except command-line options to exclude matching builds.
			buildName := pcb.Name()
			possibleBuildNames = append(possibleBuildNames, buildName)
			// -only
			if len(s.opts.Only) > 0 {
				onlyGlobs, diags := convertFilterOption(s.opts.Only, "only")
				if diags.HasErrors() {
					return nil, diags
				}
				s.config.Only = onlyGlobs
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
				onlyMatches++
			}

			// -except
			if len(s.opts.Except) > 0 {
				exceptGlobs, diags := convertFilterOption(s.opts.Except, "except")
				if diags.HasErrors() {
					return nil, diags
				}
				s.config.Except = exceptGlobs
				exclude := false
				for _, exceptGlob := range exceptGlobs {
					if exceptGlob.Match(buildName) {
						exclude = true
						break
					}
				}
				if exclude {
					exceptMatches++
					continue
				}
			}

			builder, moreDiags, generatedVars := s.startBuilder(srcUsage, s.config.EvalContext(nil))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			decoded, _ := hcl2template.DecodeHCL2Spec(srcUsage.Body, s.config.EvalContext(nil), builder)
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

			provisioners, moreDiags := s.getCoreBuildProvisioners(srcUsage, build.ProvisionerBlocks, s.config.EvalContext(variables))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			pps, moreDiags := s.getCoreBuildPostProcessors(srcUsage, build.PostProcessorsLists, s.config.EvalContext(variables), &exceptMatches)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			if build.ErrorCleanupProvisionerBlock != nil &&
				!build.ErrorCleanupProvisionerBlock.OnlyExcept.Skip(srcUsage.String()) {
				errorCleanupProv, moreDiags := s.getCoreBuildProvisioner(srcUsage, build.ErrorCleanupProvisionerBlock, s.config.EvalContext(variables))
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
	if len(s.opts.Only) > onlyMatches {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "an 'only' option was passed, but not all matches were found for the given build.",
			Detail: fmt.Sprintf("Possible build names: %v.\n"+
				"These could also be matched with a glob pattern like: 'happycloud.*'", possibleBuildNames),
		})
	}
	if len(s.opts.Except) > exceptMatches {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "an 'except' option was passed, but did not match any build.",
			Detail: fmt.Sprintf("Possible build names: %v.\n"+
				"These could also be matched with a glob pattern like: 'happycloud.*'", possibleBuildNames),
		})
	}
	return res, diags
}
