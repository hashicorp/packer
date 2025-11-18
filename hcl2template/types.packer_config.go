// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	pkrfunction "github.com/hashicorp/packer/hcl2template/function"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// PackerConfig represents a loaded Packer HCL config. It will contain
// references to all possible blocks of the allowed configuration.
type PackerConfig struct {
	Packer struct {
		VersionConstraints []VersionConstraint
		RequiredPlugins    []*RequiredPlugins
	}

	// Directory where the config files are defined
	Basedir string

	// Core Packer version, for reference by plugins and template functions.
	CorePackerVersionString string

	// directory Packer was called from
	Cwd string

	// Available Source blocks
	Sources map[SourceRef]SourceBlock

	// InputVariables and LocalVariables are the list of defined input and
	// local variables. They are of the same type but are not used in the same
	// way. Local variables will not be decoded from any config file, env var,
	// or ect. Like the Input variables will.
	InputVariables Variables
	LocalVariables Variables

	Datasources Datasources

	LocalBlocks []*LocalBlock

	ValidationOptions

	// Builds is the list of Build blocks defined in the config files.
	Builds Builds

	// HCPPackerRegistry contains the configuration for publishing the artifacts to the HCP Packer Registry.
	HCPPackerRegistry *HCPPackerRegistryBlock

	// HCPVars is the list of HCP-set variables for use later in a template
	HCPVars map[string]cty.Value

	parser *Parser
	files  []*hcl.File

	// Fields passed as command line flags
	except  []glob.Glob
	only    []glob.Glob
	force   bool
	debug   bool
	onError string
}

type ValidationOptions struct {
	WarnOnUndeclaredVar bool
}

const (
	inputVariablesAccessor = "var"
	localsAccessor         = "local"
	pathVariablesAccessor  = "path"
	sourcesAccessor        = "source"
	buildAccessor          = "build"
	packerAccessor         = "packer"
	dataAccessor           = "data"
)

type BlockContext int

const (
	InputVariableContext BlockContext = iota
	LocalContext
	BuildContext
	DatasourceContext
	NilContext
)

// EvalContext returns the *hcl.EvalContext that will be passed to an hcl
// decoder in order to tell what is the actual value of a var or a local and
// the list of defined functions.
func (cfg *PackerConfig) EvalContext(ctx BlockContext, variables map[string]cty.Value) *hcl.EvalContext {
	inputVariables := cfg.InputVariables.Values()
	localVariables := cfg.LocalVariables.Values()
	ectx := &hcl.EvalContext{
		Functions: Functions(cfg.Basedir),
		Variables: map[string]cty.Value{
			inputVariablesAccessor: cty.ObjectVal(inputVariables),
			localsAccessor:         cty.ObjectVal(localVariables),
			sourcesAccessor: cty.ObjectVal(map[string]cty.Value{
				"type": cty.UnknownVal(cty.String),
				"name": cty.UnknownVal(cty.String),
			}),
			buildAccessor: cty.UnknownVal(cty.EmptyObject),
			pathVariablesAccessor: cty.ObjectVal(map[string]cty.Value{
				"cwd":  cty.StringVal(strings.ReplaceAll(cfg.Cwd, `\`, `/`)),
				"root": cty.StringVal(strings.ReplaceAll(cfg.Basedir, `\`, `/`)),
			}),
		},
	}

	packerVars := map[string]cty.Value{
		"version":            cty.StringVal(cfg.CorePackerVersionString),
		"iterationID":        cty.UnknownVal(cty.String),
		"versionFingerprint": cty.UnknownVal(cty.String),
	}

	iterID, ok := cfg.HCPVars["iterationID"]
	if ok {
		packerVars["iterationID"] = iterID
	}
	versionFP, ok := cfg.HCPVars["versionFingerprint"]
	if ok {
		packerVars["versionFingerprint"] = versionFP
	}

	ectx.Variables[packerAccessor] = cty.ObjectVal(packerVars)

	// In the future we'd like to load and execute HCL blocks using a graph
	// dependency tree, so that any block can use any block whatever the
	// order.
	// For now, don't add DataSources if there's a NilContext, which gets
	// used with packer console.
	switch ctx {
	case LocalContext, BuildContext, DatasourceContext:
		datasourceVariables, _ := cfg.Datasources.Values()
		ectx.Variables[dataAccessor] = cty.ObjectVal(datasourceVariables)
	}

	for k, v := range variables {
		ectx.Variables[k] = v
	}
	return ectx
}

// decodeInputVariables looks in the found blocks for 'variables' and
// 'variable' blocks. It should be called firsthand so that other blocks can
// use the variables.
func (c *PackerConfig) decodeInputVariables(f *hcl.File) hcl.Diagnostics {
	var diags hcl.Diagnostics

	content, _ := f.Body.Content(configSchema)

	// for input variables we allow to use env in the default value section.
	ectx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"env": pkrfunction.EnvFunc,
		},
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case variableLabel:
			moreDiags := c.InputVariables.decodeVariableBlock(block, ectx)
			diags = append(diags, moreDiags...)
		case variablesLabel:
			attrs, moreDiags := block.Body.JustAttributes()
			diags = append(diags, moreDiags...)
			for key, attr := range attrs {
				moreDiags = c.InputVariables.decodeVariable(key, attr, ectx)
				diags = append(diags, moreDiags...)
			}
		}
	}
	return diags
}

// parseLocalVariableBlocks looks in the AST for 'local' and 'locals' blocks and
// returns them all.
func parseLocalVariableBlocks(f *hcl.File) ([]*LocalBlock, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, _ := f.Body.Content(configSchema)

	var locals []*LocalBlock

	for _, block := range content.Blocks {
		switch block.Type {
		case localLabel:
			block, moreDiags := decodeLocalBlock(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				return locals, diags
			}
			locals = append(locals, block)
		case localsLabel:
			attrs, moreDiags := block.Body.JustAttributes()
			diags = append(diags, moreDiags...)
			for name, attr := range attrs {
				locals = append(locals, &LocalBlock{
					LocalName: name,
					Expr:      attr.Expr,
				})
			}
		}
	}

	return locals, diags
}

// unused maybe remove?
// func (c *PackerConfig) localByName(local string) (*LocalBlock, error) {
// 	for _, loc := range c.LocalBlocks {
// 		if loc.LocalName != local {
// 			continue
// 		}
//
// 		return loc, nil
// 	}
//
// 	return nil, fmt.Errorf("local %s not found", local)
// }

func (c *PackerConfig) evaluateLocalVariables(locals []*LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if len(locals) == 0 {
		return diags
	}

	if c.LocalVariables == nil {
		c.LocalVariables = Variables{}
	}

	for _, local := range c.LocalBlocks {
		// Note: when looking at the expressions, we only need to care about
		// attributes, as HCL2 expressions are not allowed in a block's labels.
		vars := FilterTraversalsByType(local.Expr.Variables(), "local")

		var localDeps []refString
		for _, v := range vars {
			// Some local variables may be locally aliased as
			// `local`, which
			if len(v) < 2 {
				continue
			}

			depRef, err := NewRefStringFromDep(v)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to extract dependency name from traversal ref",
					Detail: fmt.Sprintf("while preparing for evaluation of local variable %q, "+
						"a dependency was unable to be converted to a refString. "+
						"This is likely a Packer bug, please consider reporting it.", local.Name()),
				})
				continue
			}

			localDeps = append(localDeps, depRef)
		}
		local.dependencies = localDeps
	}

	// Immediately return in case the dependencies couldn't be figured out.
	if diags.HasErrors() {
		return diags
	}

	for _, local := range c.LocalBlocks {
		diags = diags.Extend(c.recursivelyEvaluateLocalVariable(local, 0))
	}

	return diags
}

// checkForDuplicateLocalDefinition walks through the list of defined variables
// in order to detect duplicate locals definitions.
func (c *PackerConfig) checkForDuplicateLocalDefinition() hcl.Diagnostics {
	var diags hcl.Diagnostics

	localNames := map[string]*LocalBlock{}

	for _, block := range c.LocalBlocks {
		loc, ok := localNames[block.LocalName]
		if ok {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate local definition",
				Detail: fmt.Sprintf("Local variable %q is defined twice in your templates. Other definition found at %q",
					block.LocalName, loc.Expr.Range()),
				Subject: block.Expr.Range().Ptr(),
			})
			continue
		}

		localNames[block.LocalName] = block
	}

	return diags
}

func (c *PackerConfig) recursivelyEvaluateLocalVariable(local *LocalBlock, depth int) hcl.Diagnostics {
	// If the variable already was evaluated, we can return immediately
	if local.evaluated {
		return nil
	}

	if depth >= 10 {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Max local recursion depth exceeded.",
			Detail: "An error occured while recursively evaluating locals." +
				"Your local variables likely have a cyclic dependency. " +
				"Please simplify your config to continue. ",
		}}
	}

	var diags hcl.Diagnostics

	for _, dep := range local.dependencies {
		locBlock, err := c.getComponentByRef(dep)
		if err != nil {
			return diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to get local variable",
				Detail: fmt.Sprintf("While evaluating %q, its dependency %q was not found, is it defined?",
					local.Name(), dep.String()),
			})
		}

		localDiags := c.recursivelyEvaluateLocalVariable(locBlock.(*LocalBlock), depth+1)
		diags = diags.Extend(localDiags)
	}

	val, locDiags := c.evaluateLocalVariable(local)
	if !locDiags.HasErrors() {
		c.LocalVariables[local.LocalName] = val
	}

	return diags.Extend(locDiags)
}

func (cfg *PackerConfig) evaluateLocalVariable(local *LocalBlock) (*Variable, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	value, moreDiags := local.Expr.Value(cfg.EvalContext(LocalContext, nil))

	local.evaluated = true

	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags
	}
	return &Variable{
		Name:      local.LocalName,
		Sensitive: local.Sensitive,
		Values: []VariableAssignment{{
			Value: value,
			Expr:  local.Expr,
			From:  "default",
		}},
		Type: value.Type(),
	}, diags
}

func (cfg *PackerConfig) evaluateDatasources(skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	dependencies := map[DatasourceRef][]DatasourceRef{}
	for ref, ds := range cfg.Datasources {
		if ds.value != (cty.Value{}) {
			continue
		}
		// Pre-examine body of this data source to see if it uses another data
		// source in any of its input expressions. If so, skip evaluating it for
		// now, and add it to a list of datasources to evaluate again, later,
		// with the datasources in its context.
		dependencies[ref] = []DatasourceRef{}

		// Note: when looking at the expressions, we only need to care about
		// attributes, as HCL2 expressions are not allowed in a block's labels.
		vars := GetVarsByType(ds.block, "data")
		for _, v := range vars {
			// construct, backwards, the data source type and name we
			// need to evaluate before this one can be evaluated.
			dependsOn := DatasourceRef{
				Type: v[1].(hcl.TraverseAttr).Name,
				Name: v[2].(hcl.TraverseAttr).Name,
			}
			dependencies[ref] = append(dependencies[ref], dependsOn)
		}
	}

	// Now that most of our data sources have been started and executed, we can
	// try to execute the ones that depend on other data sources.
	for ref := range dependencies {
		_, moreDiags := cfg.recursivelyEvaluateDatasources(ref, dependencies, skipExecution, 0)
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

func (cfg *PackerConfig) recursivelyEvaluateDatasources(ref DatasourceRef, dependencies map[DatasourceRef][]DatasourceRef, skipExecution bool, depth int) (map[DatasourceRef][]DatasourceRef, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var moreDiags hcl.Diagnostics

	if depth > 10 {
		// Add a comment about recursion.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Max datasource recursion depth exceeded.",
			Detail: "An error occured while recursively evaluating data " +
				"sources. Either your data source depends on more than ten " +
				"other data sources, or your data sources have a cyclic " +
				"dependency. Please simplify your config to continue. ",
			Subject: &(cfg.Datasources[ref]).block.DefRange,
		})
		return dependencies, diags
	}

	ds := cfg.Datasources[ref]
	// Make sure everything ref depends on has already been evaluated.
	for _, dep := range dependencies[ref] {
		if _, ok := dependencies[dep]; ok {
			depth += 1
			// If this dependency is not in the map, it means we've already
			// launched and executed this datasource. Otherwise, it means
			// we still need to run it. RECURSION TIME!!
			dependencies, moreDiags = cfg.recursivelyEvaluateDatasources(dep, dependencies, skipExecution, depth)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				return dependencies, diags
			}
		}
	}
	// If we've gotten here, then it means ref doesn't seem to have any further
	// dependencies we need to evaluate first. Evaluate it, with the cfg's full
	// data source context.
	datasource, startDiags := cfg.startDatasource(ds)
	if startDiags.HasErrors() {
		diags = append(diags, startDiags...)
		return dependencies, diags
	}

	if skipExecution {
		placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
		ds.value = placeholderValue
		cfg.Datasources[ref] = ds
		return dependencies, diags
	}

	opts, _ := decodeHCL2Spec(ds.block.Body, cfg.EvalContext(DatasourceContext, nil), datasource)
	sp := packer.CheckpointReporter.AddSpan(ref.Type, "datasource", opts)
	realValue, err := datasource.Execute()
	sp.End(err)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &cfg.Datasources[ref].block.DefRange,
			Severity: hcl.DiagError,
		})
		return dependencies, diags
	}

	ds.value = realValue
	cfg.Datasources[ref] = ds
	// remove ref from the dependencies map.
	delete(dependencies, ref)
	return dependencies, diags
}

func (cfg *PackerConfig) evaluateDatasource(ds DatasourceBlock, skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// If we've gotten here, then it means ref doesn't seem to have any further
	// dependencies we need to evaluate first. Evaluate it, with the cfg's full
	// data source context.
	datasource, startDiags := cfg.startDatasource(ds)
	if startDiags.HasErrors() {
		diags = append(diags, startDiags...)
		return diags
	}

	if skipExecution {
		placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
		ds.value = placeholderValue
		cfg.Datasources[ds.Ref()] = ds
		return diags
	}

	opts, _ := decodeHCL2Spec(ds.block.Body, cfg.EvalContext(DatasourceContext, nil), datasource)
	sp := packer.CheckpointReporter.AddSpan(ds.Ref().Type, "datasource", opts)
	realValue, err := datasource.Execute()
	sp.End(err)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &cfg.Datasources[ds.Ref()].block.DefRange,
			Severity: hcl.DiagError,
		})
		return diags
	}

	ds.value = realValue
	cfg.Datasources[ds.Ref()] = ds

	return diags
}

// getCoreBuildProvisioners takes a list of provisioner block, starts according
// provisioners and sends parsed HCL2 over to it.
func (cfg *PackerConfig) getCoreBuildProvisioners(source SourceUseBlock, blocks []*ProvisionerBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		if pb.OnlyExcept.Skip(source.String()) {
			continue
		}

		coreBuildProv, moreDiags := cfg.getCoreBuildProvisioner(source, pb, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, coreBuildProv)
	}
	return res, diags
}

func (cfg *PackerConfig) getCoreBuildProvisioner(source SourceUseBlock, pb *ProvisionerBlock, ectx *hcl.EvalContext) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	provisioner, moreDiags := cfg.startProvisioner(source, pb, ectx)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return packer.CoreBuildProvisioner{}, diags
	}

	flatProvisionerCfg, _ := decodeHCL2Spec(pb.HCL2Ref.Rest, ectx, provisioner)

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

	if pb.PType == "hcp-sbom" {
		provisioner = &packer.SBOMInternalProvisioner{
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

// getCoreBuildProvisioners takes a list of post processor block, starts
// according provisioners and sends parsed HCL2 over to it.
func (cfg *PackerConfig) getCoreBuildPostProcessors(source SourceUseBlock, blocksList [][]*PostProcessorBlock, ectx *hcl.EvalContext, exceptMatches *int) ([][]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
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
			for _, exceptGlob := range cfg.except {
				if exceptGlob.Match(name) {
					exclude = true
					*exceptMatches = *exceptMatches + 1
					break
				}
			}
			if exclude {
				break
			}

			postProcessor, moreDiags := cfg.startPostProcessor(source, ppb, ectx)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			flatPostProcessorCfg, moreDiags := decodeHCL2Spec(ppb.HCL2Ref.Rest, ectx, postProcessor)

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

// GetHCPPackerRegistryBlock return the HCP registry configuration block
// that can should be used for the current build. Right now, it should
// use the block at the top level but support the block inside the first
// build block with a deprecation diagnostic
func (cfg *PackerConfig) GetHCPPackerRegistryBlock() (*HCPPackerRegistryBlock, hcl.Diagnostics) {
	var block *HCPPackerRegistryBlock
	var diags hcl.Diagnostics

	multipleRegistryDiag := func(block *HCPPackerRegistryBlock) *hcl.Diagnostic {
		return &hcl.Diagnostic{
			Summary:  "Multiple HCP Packer registry block declaration",
			Subject:  block.HCL2Ref.DefRange.Ptr(),
			Severity: hcl.DiagError,
			Detail: "Multiple " + buildHCPPackerRegistryLabel + " blocks have been found, only one can be defined " +
				"in HCL2 templates. Starting with Packer 1.12.1, it is recommended to move it to the " +
				"top-level configuration instead of within a build block.",
		}
	}
	// We start by looking in the build blocks
	for _, build := range cfg.Builds {
		if build.HCPPackerRegistry != nil {
			if block != nil {
				// error multiple build block
				diags = diags.Append(multipleRegistryDiag(build.HCPPackerRegistry))
				continue
			}
			block = build.HCPPackerRegistry
			diags = diags.Append(&hcl.Diagnostic{
				Summary:  "Build block level " + buildHCPPackerRegistryLabel + " are deprecated",
				Subject:  &block.DefRange,
				Severity: hcl.DiagWarning,
				Detail: "Starting with Packer 1.12.1, it is recommended to move it to the " +
					"top-level configuration instead of within a build block.",
			})
		}
	}

	if block != nil && cfg.HCPPackerRegistry != nil {
		diags = diags.Append(multipleRegistryDiag(block))
	}

	if cfg.HCPPackerRegistry != nil {
		block = cfg.HCPPackerRegistry
	}

	return block, diags
}

// GetBuilds returns a list of packer Build based on the HCL2 parsed build
// blocks. All Builders, Provisioners and Post Processors will be started and
// configured.
func (cfg *PackerConfig) GetBuilds(opts packer.GetBuildsOptions) ([]*packer.CoreBuild, hcl.Diagnostics) {
	res := []*packer.CoreBuild{}
	var diags hcl.Diagnostics
	possibleBuildNames := []string{}

	cfg.debug = opts.Debug
	cfg.force = opts.Force
	cfg.onError = opts.OnError

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
					Summary:  "Unknown " + sourceLabel + " " + srcUsage.String(),
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

			pcb.SetDebug(cfg.debug)
			pcb.SetForce(cfg.force)
			pcb.SetOnError(cfg.onError)

			// Apply the -only and -except command-line options to exclude matching builds.
			buildName := pcb.Name()
			possibleBuildNames = append(possibleBuildNames, buildName)
			// -only
			if len(opts.Only) > 0 {
				onlyGlobs, diags := convertFilterOption(opts.Only, "only")
				if diags.HasErrors() {
					return nil, diags
				}
				cfg.only = onlyGlobs
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
				cfg.except = exceptGlobs
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

			builder, moreDiags, generatedVars := cfg.startBuilder(srcUsage, cfg.EvalContext(BuildContext, nil))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			decoded, _ := decodeHCL2Spec(srcUsage.Body, cfg.EvalContext(BuildContext, nil), builder)
			pcb.HCLConfig = decoded
			pcb.BuilderType = srcUsage.Type

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
				sourcesAccessor: cty.ObjectVal(srcUsage.ctyValues()),
				buildAccessor:   cty.ObjectVal(unknownBuildValues),
			}

			provisioners, moreDiags := cfg.getCoreBuildProvisioners(srcUsage, build.ProvisionerBlocks, cfg.EvalContext(BuildContext, variables))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			pps, moreDiags := cfg.getCoreBuildPostProcessors(srcUsage, build.PostProcessorsLists, cfg.EvalContext(BuildContext, variables), &opts.ExceptMatches)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			if build.ErrorCleanupProvisionerBlock != nil &&
				!build.ErrorCleanupProvisionerBlock.OnlyExcept.Skip(srcUsage.String()) {
				errorCleanupProv, moreDiags := cfg.getCoreBuildProvisioner(srcUsage, build.ErrorCleanupProvisionerBlock, cfg.EvalContext(BuildContext, variables))
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

			pcb.SensitiveVars = make([]string, 0, len(cfg.InputVariables))

			for key, variable := range cfg.InputVariables {
				if variable.Sensitive {
					pcb.SensitiveVars = append(pcb.SensitiveVars, key)
				}
			}

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

var PackerConsoleHelp = strings.TrimSpace(`
Packer console HCL2 Mode.
The Packer console allows you to experiment with Packer interpolations.
You may access variables and functions in the Packer config you called the
console with.

Type in the interpolation to test and hit <enter> to see the result.

"upper(var.foo.id)" would evaluate to the ID of "foo" and uppercase is, if it
exists in your config file.

"variables" will dump all available variables and their values.

To exit the console, type "exit" and hit <enter>, or use Control-C.

/!\ It is not possible to use go templating interpolation like "{{timestamp}}"
with in HCL2 mode.
`)

func (p *PackerConfig) EvaluateExpression(line string) (out string, exit bool, diags hcl.Diagnostics) {
	switch {
	case line == "":
		return "", false, nil
	case line == "exit":
		return "", true, nil
	case line == "help":
		return PackerConsoleHelp, false, nil
	case line == "variables":
		return p.printVariables(), false, nil
	default:
		return p.handleEval(line)
	}
}

func (p *PackerConfig) printVariables() string {
	out := &strings.Builder{}
	out.WriteString("> input-variables:\n\n")
	keys := p.InputVariables.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		v := p.InputVariables[key]
		val := v.Value()
		fmt.Fprintf(out, "var.%s: %q\n", v.Name, PrintableCtyValue(val))
	}
	out.WriteString("\n> local-variables:\n\n")
	keys = p.LocalVariables.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		v := p.LocalVariables[key]
		val := v.Value()
		fmt.Fprintf(out, "local.%s: %q\n", v.Name, PrintableCtyValue(val))
	}
	return out.String()
}

func (p *PackerConfig) printBuilds() string {
	out := &strings.Builder{}
	out.WriteString("> builds:\n")
	for i, build := range p.Builds {
		name := build.Name
		if name == "" {
			name = fmt.Sprintf("<unnamed build %d>", i)
		}
		fmt.Fprintf(out, "\n  > %s:\n", name)
		if build.Description != "" {
			fmt.Fprintf(out, "\n  > Description: %s\n", build.Description)
		}
		fmt.Fprintf(out, "\n    sources:\n")
		if len(build.Sources) == 0 {
			fmt.Fprintf(out, "\n      <no source>\n")
		}
		for _, source := range build.Sources {
			fmt.Fprintf(out, "\n      %s\n", source.String())
		}
		fmt.Fprintf(out, "\n    provisioners:\n\n")
		if len(build.ProvisionerBlocks) == 0 {
			fmt.Fprintf(out, "      <no provisioner>\n")
		}
		for _, prov := range build.ProvisionerBlocks {
			str := prov.PType
			if prov.PName != "" {
				str = strings.Join([]string{prov.PType, prov.PName}, ".")
			}
			fmt.Fprintf(out, "      %s\n", str)
		}
		fmt.Fprintf(out, "\n    post-processors:\n")
		if len(build.PostProcessorsLists) == 0 {
			fmt.Fprintf(out, "\n      <no post-processor>\n")
		}
		for i, ppList := range build.PostProcessorsLists {
			fmt.Fprintf(out, "\n      %d:\n", i)
			for _, pp := range ppList {
				str := pp.PType
				if pp.PName != "" {
					str = strings.Join([]string{pp.PType, pp.PName}, ".")
				}
				fmt.Fprintf(out, "        %s\n", str)
			}
		}
	}
	return out.String()
}

func (p *PackerConfig) handleEval(line string) (out string, exit bool, diags hcl.Diagnostics) {

	// Parse the given line as an expression
	expr, parseDiags := hclsyntax.ParseExpression([]byte(line), "<console-input>", hcl.Pos{Line: 1, Column: 1})
	diags = append(diags, parseDiags...)
	if parseDiags.HasErrors() {
		return "", false, diags
	}

	val, valueDiags := expr.Value(p.EvalContext(NilContext, nil))
	diags = append(diags, valueDiags...)
	if valueDiags.HasErrors() {
		return "", false, diags
	}

	return PrintableCtyValue(val), false, diags
}

func (p *PackerConfig) FixConfig(_ packer.FixConfigOptions) (diags hcl.Diagnostics) {
	// No Fixers exist for HCL2 configs so there is nothing to do here for now.
	return
}

func (p *PackerConfig) InspectConfig(opts packer.InspectConfigOptions) int {

	ui := opts.Ui
	ui.Say("Packer Inspect: HCL2 mode\n")
	ui.Say(p.printVariables())
	ui.Say(p.printBuilds())
	return 0
}
