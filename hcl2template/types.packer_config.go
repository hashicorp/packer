// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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

	// HCPVars is the list of HCP-set variables for use later in a template
	HCPVars map[string]cty.Value

	parser *Parser
	files  map[string]*hcl.File

	// Fields passed as command line flags
	Force   bool
	Debug   bool
	OnError string

	// except/only are options to filter builds and post-processors
	//
	// if an option is specified but unused, we print a warning, stating
	// which option is specified, but unused
	Except     map[string]glob.Glob
	Only       map[string]glob.Glob
	ExceptUses map[string]bool
	OnlyUses   map[string]bool
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
			packerAccessor: cty.ObjectVal(map[string]cty.Value{
				"version":     cty.StringVal(cfg.CorePackerVersionString),
				"iterationID": cty.UnknownVal(cty.String),
			}),
			pathVariablesAccessor: cty.ObjectVal(map[string]cty.Value{
				"cwd":  cty.StringVal(strings.ReplaceAll(cfg.Cwd, `\`, `/`)),
				"root": cty.StringVal(strings.ReplaceAll(cfg.Basedir, `\`, `/`)),
			}),
		},
	}

	iterID, ok := cfg.HCPVars["iterationID"]
	if ok {
		ectx.Variables[packerAccessor] = cty.ObjectVal(map[string]cty.Value{
			"version":     cty.StringVal(cfg.CorePackerVersionString),
			"iterationID": iterID,
		})
	}

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

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)

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

func (cfg *PackerConfig) CheckForDuplicateLocalDefinition() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// we could sort by name and then check contiguous names to use less memory,
	// but using a map sounds good enough.
	names := map[string]struct{}{}
	for _, local := range cfg.LocalBlocks {
		if _, found := names[local.Name]; found {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate local definition",
				Detail:   "Duplicate " + local.Name + " definition found.",
				Subject:  local.Expr.Range().Ptr(),
			})
			continue
		}
		names[local.Name] = struct{}{}
	}
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

	return packer.CoreBuildProvisioner{
		PType:       pb.PType,
		PName:       pb.PName,
		Provisioner: provisioner,
		HCLConfig:   flatProvisionerCfg,
	}, diags
}

// getCoreBuildProvisioners takes a list of post processor block, starts
// according provisioners and sends parsed HCL2 over to it.
func (cfg *PackerConfig) getCoreBuildPostProcessors(source SourceUseBlock, blocksList [][]*PostProcessorBlock, ectx *hcl.EvalContext) ([][]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
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
			for p, exceptGlob := range cfg.Except {
				if exceptGlob.Match(name) {
					exclude = true
					cfg.ExceptUses[p] = true
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

func (cfg *PackerConfig) PrepareGlobUsage() {
	if cfg.OnlyUses == nil {
		cfg.OnlyUses = map[string]bool{}
	}
	for only := range cfg.Only {
		cfg.OnlyUses[only] = false
	}

	if cfg.ExceptUses == nil {
		cfg.ExceptUses = map[string]bool{}
	}
	for except := range cfg.Except {
		cfg.ExceptUses[except] = false
	}
}

func (cfg *PackerConfig) ReportUnusedFilters(buildNames []string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	onlyUnused := getUnusedFilters(cfg.OnlyUses)
	if onlyUnused != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Unused --only filters specified",
			Detail: fmt.Sprintf(
				"Some --only options were specified in the command-line, but weren't used to get builds from the template: %v\n"+
					"List of available builds: %v", onlyUnused, buildNames),
		})
	}

	exceptUnused := getUnusedFilters(cfg.ExceptUses)
	if exceptUnused != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Unused --except filters specified",
			Detail: fmt.Sprintf(
				"Some --except options were specified in the command-line, but weren't used to filter out builds or post-processors from the template: %v\n"+
					"List of available builds: %v", exceptUnused, buildNames),
		})
	}

	return diags
}

func (cfg *PackerConfig) Files() map[string]*hcl.File {
	return cfg.files
}

func (cfg *PackerConfig) GetBuilds(opts packer.GetBuildsOptions) ([]packersdk.Build, hcl.Diagnostics) {
	var allBuilds []packersdk.Build
	var diags hcl.Diagnostics

	if len(cfg.Builds) == 0 {
		return nil, append(diags, &hcl.Diagnostic{
			Summary:  "Missing build block",
			Detail:   "A build block with one or more sources is required for executing a build.",
			Severity: hcl.DiagError,
		})
	}

	var convertDiags hcl.Diagnostics
	cfg.Debug = opts.Debug
	cfg.Except, convertDiags = ConvertFilterOption(opts.Except, "except")
	diags = diags.Extend(convertDiags)
	cfg.Only, convertDiags = ConvertFilterOption(opts.Only, "only")
	diags = diags.Extend(convertDiags)
	cfg.Force = opts.Force
	cfg.OnError = opts.OnError

	cfg.PrepareGlobUsage()

	for _, build := range cfg.Builds {
		cbs, cbDiags := build.ToCoreBuilds(cfg)
		diags = diags.Extend(cbDiags)

		for _, cb := range cbs {
			cb.SetDebug(opts.Debug)
			cb.SetForce(opts.Force)
			cb.SetOnError(opts.OnError)

			cb.Prepared = true

			// Prepare just sets the "prepareCalled" flag on CoreBuild, since
			// we did all the prep here.
			_, err := cb.Prepare()
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Preparing packer core build %s failed", cb.Name()),
					Detail:   err.Error(),
				})
			}

			allBuilds = append(allBuilds, cb)
		}
	}

	buildNames := []string{}
	for _, cb := range allBuilds {
		buildNames = append(buildNames, cb.Name())
	}

	diags = diags.Extend(
		cfg.ReportUnusedFilters(
			buildNames,
		),
	)

	return allBuilds, diags
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

func (cfg *PackerConfig) Initialize(opts packer.InitializeOptions) hcl.Diagnostics {
	diags := cfg.InputVariables.ValidateValues()
	diags = diags.Extend(cfg.CheckForDuplicateLocalDefinition())
	diags = diags.Extend(cfg.executeDatasources(opts.SkipDatasourcesExecution))
	diags = diags.Extend(cfg.evaluateLocalVariables())

	filterVarsFromLogs(cfg.InputVariables)
	filterVarsFromLogs(cfg.LocalVariables)

	diags = diags.Extend(cfg.initializeBlocks())

	return diags
}
