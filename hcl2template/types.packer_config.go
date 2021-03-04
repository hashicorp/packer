package hcl2template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
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
	Strict bool
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
	inputVariables, _ := cfg.InputVariables.Values()
	localVariables, _ := cfg.LocalVariables.Values()
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
				"version": cty.StringVal(cfg.CorePackerVersionString),
			}),
			pathVariablesAccessor: cty.ObjectVal(map[string]cty.Value{
				"cwd":  cty.StringVal(strings.ReplaceAll(cfg.Cwd, `\`, `/`)),
				"root": cty.StringVal(strings.ReplaceAll(cfg.Basedir, `\`, `/`)),
			}),
		},
	}

	// Currently the places where you can make references to other blocks
	// from one is very 'procedural', and in this specific case, we could make
	// the data sources available to other datasources, but this would be
	// order dependant, meaning that if you define two datasources in two
	// different blocks, the second one can use the first one, but not the
	// other way around; which would be totally confusing; so - for now -
	// datasources can't use other datasources.
	// In the future we'd like to load and execute HCL blocks using a graph
	// dependency tree, so that any block can use any block whatever the
	// order.
	switch ctx {
	case LocalContext, BuildContext:
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

// parseLocalVariables looks in the found blocks for 'locals' blocks. It
// should be called after parsing input variables so that they can be
// referenced.
func (c *PackerConfig) parseLocalVariables(f *hcl.File) ([]*LocalBlock, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)

	locals := c.LocalBlocks

	for _, block := range content.Blocks {
		switch block.Type {
		case localLabel:
			l, moreDiags := decodeLocalBlock(block, locals)
			diags = append(diags, moreDiags...)
			if l != nil {
				locals = append(locals, l)
			}
			if moreDiags.HasErrors() {
				return locals, diags
			}
		case localsLabel:
			attrs, moreDiags := block.Body.JustAttributes()
			diags = append(diags, moreDiags...)
			for name, attr := range attrs {
				if _, found := c.LocalVariables[name]; found {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Duplicate value in " + localsLabel,
						Detail:   "Duplicate " + name + " definition found.",
						Subject:  attr.NameRange.Ptr(),
						Context:  block.DefRange.Ptr(),
					})
					return locals, diags
				}
				locals = append(locals, &LocalBlock{
					Name: name,
					Expr: attr.Expr,
				})
			}
		}
	}

	c.LocalBlocks = locals
	return locals, diags
}

func (c *PackerConfig) evaluateLocalVariables(locals []*LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if len(locals) > 0 && c.LocalVariables == nil {
		c.LocalVariables = Variables{}
	}

	var retry, previousL int
	for len(locals) > 0 {
		local := locals[0]
		moreDiags := c.evaluateLocalVariable(local)
		if moreDiags.HasErrors() {
			if len(locals) == 1 {
				// If this is the only local left there's no need
				// to try evaluating again
				return append(diags, moreDiags...)
			}
			if previousL == len(locals) {
				if retry == 100 {
					// To get to this point, locals must have a circle dependency
					return append(diags, moreDiags...)
				}
				retry++
			}
			previousL = len(locals)

			// If local uses another local that has not been evaluated yet this could be the reason of errors
			// Push local to the end of slice to be evaluated later
			locals = append(locals, local)
		} else {
			retry = 0
			diags = append(diags, moreDiags...)
		}
		// Remove local from slice
		locals = append(locals[:0], locals[1:]...)
	}

	return diags
}

func (c *PackerConfig) evaluateLocalVariable(local *LocalBlock) hcl.Diagnostics {
	var diags hcl.Diagnostics
	value, moreDiags := local.Expr.Value(c.EvalContext(LocalContext, nil))
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return diags
	}
	c.LocalVariables[local.Name] = &Variable{
		Name:      local.Name,
		Sensitive: local.Sensitive,
		Values: []VariableAssignment{{
			Value: value,
			Expr:  local.Expr,
			From:  "default",
		}},
		Type: value.Type(),
	}

	return diags
}

func (cfg *PackerConfig) evaluateDatasources(skipExecution bool) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for ref, ds := range cfg.Datasources {
		if ds.value != (cty.Value{}) {
			continue
		}

		datasource, startDiags := cfg.startDatasource(cfg.parser.PluginConfig.DataSources, ref)
		diags = append(diags, startDiags...)
		if diags.HasErrors() {
			continue
		}

		if skipExecution {
			placeholderValue := cty.UnknownVal(hcldec.ImpliedType(datasource.OutputSpec()))
			ds.value = placeholderValue
			cfg.Datasources[ref] = ds
			continue
		}

		realValue, err := datasource.Execute()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  err.Error(),
				Subject:  &cfg.Datasources[ref].block.DefRange,
				Severity: hcl.DiagError,
			})
			continue
		}
		ds.value = realValue
		cfg.Datasources[ref] = ds
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
			for _, exceptGlob := range cfg.except {
				if exceptGlob.Match(name) {
					exclude = true
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
			pps = append(pps, packer.CoreBuildPostProcessor{
				PostProcessor:     postProcessor,
				PName:             ppb.PName,
				PType:             ppb.PType,
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
func (cfg *PackerConfig) GetBuilds(opts packer.GetBuildsOptions) ([]packersdk.Build, hcl.Diagnostics) {
	res := []packersdk.Build{}
	var diags hcl.Diagnostics

	cfg.debug = opts.Debug
	cfg.force = opts.Force
	cfg.onError = opts.OnError

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

			// Apply the -only and -except command-line options to exclude matching builds.
			buildName := pcb.Name()
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
					continue
				}
			}

			builder, moreDiags, generatedVars := cfg.startBuilder(srcUsage, cfg.EvalContext(BuildContext, nil))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

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
			pps, moreDiags := cfg.getCoreBuildPostProcessors(srcUsage, build.PostProcessorsLists, cfg.EvalContext(BuildContext, variables))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			if build.ErrorCleanupProvisionerBlock != nil {
				if !build.ErrorCleanupProvisionerBlock.OnlyExcept.Skip(srcUsage.String()) {
					errorCleanupProv, moreDiags := cfg.getCoreBuildProvisioner(srcUsage, build.ErrorCleanupProvisionerBlock, cfg.EvalContext(BuildContext, variables))
					diags = append(diags, moreDiags...)
					if moreDiags.HasErrors() {
						continue
					}
					pcb.CleanupProvisioner = errorCleanupProv
				}
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
		val, _ := v.Value()
		fmt.Fprintf(out, "var.%s: %q\n", v.Name, PrintableCtyValue(val))
	}
	out.WriteString("\n> local-variables:\n\n")
	keys = p.LocalVariables.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		v := p.LocalVariables[key]
		val, _ := v.Value()
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
