package hcl2template

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// PackerConfig represents a loaded Packer HCL config. It will contain
// references to all possible blocks of the allowed configuration.
type PackerConfig struct {
	// Directory where the config files are defined
	Basedir string
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

	ValidationOptions

	// Builds is the list of Build blocks defined in the config files.
	Builds Builds

	builderSchemas packer.BuilderStore

	provisionersSchemas packer.ProvisionerStore

	postProcessorsSchemas packer.PostProcessorStore

	except []glob.Glob
	only   []glob.Glob
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
)

// EvalContext returns the *hcl.EvalContext that will be passed to an hcl
// decoder in order to tell what is the actual value of a var or a local and
// the list of defined functions.
func (cfg *PackerConfig) EvalContext(variables map[string]cty.Value) *hcl.EvalContext {
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
			pathVariablesAccessor: cty.ObjectVal(map[string]cty.Value{
				"cwd":  cty.StringVal(strings.ReplaceAll(cfg.Cwd, `\`, `/`)),
				"root": cty.StringVal(strings.ReplaceAll(cfg.Basedir, `\`, `/`)),
			}),
		},
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

	for _, block := range content.Blocks {
		switch block.Type {
		case variableLabel:
			moreDiags := c.InputVariables.decodeVariableBlock(block, nil)
			diags = append(diags, moreDiags...)
		case variablesLabel:
			attrs, moreDiags := block.Body.JustAttributes()
			diags = append(diags, moreDiags...)
			for key, attr := range attrs {
				moreDiags = c.InputVariables.decodeVariable(key, attr, nil)
				diags = append(diags, moreDiags...)
			}
		}
	}
	return diags
}

// parseLocalVariables looks in the found blocks for 'locals' blocks. It
// should be called after parsing input variables so that they can be
// referenced.
func (c *PackerConfig) parseLocalVariables(f *hcl.File) ([]*Local, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)
	var locals []*Local

	for _, block := range content.Blocks {
		switch block.Type {
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
					return nil, diags
				}
				locals = append(locals, &Local{
					Name: name,
					Expr: attr.Expr,
				})
			}
		}
	}

	return locals, diags
}

func (c *PackerConfig) evaluateLocalVariables(locals []*Local) hcl.Diagnostics {
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

func (c *PackerConfig) evaluateLocalVariable(local *Local) hcl.Diagnostics {
	var diags hcl.Diagnostics

	value, moreDiags := local.Expr.Value(c.EvalContext(nil))
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return diags
	}
	c.LocalVariables[local.Name] = &Variable{
		Name:         local.Name,
		DefaultValue: value,
		Type:         value.Type(),
	}

	return diags
}

// getCoreBuildProvisioners takes a list of provisioner block, starts according
// provisioners and sends parsed HCL2 over to it.
func (cfg *PackerConfig) getCoreBuildProvisioners(source SourceBlock, blocks []*ProvisionerBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		if pb.OnlyExcept.Skip(source.String()) {
			continue
		}
		provisioner, moreDiags := cfg.startProvisioner(source, pb, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
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

		res = append(res, packer.CoreBuildProvisioner{
			PType:       pb.PType,
			PName:       pb.PName,
			Provisioner: provisioner,
		})
	}
	return res, diags
}

// getCoreBuildProvisioners takes a list of post processor block, starts
// according provisioners and sends parsed HCL2 over to it.
func (cfg *PackerConfig) getCoreBuildPostProcessors(source SourceBlock, blocks []*PostProcessorBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildPostProcessor{}
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
		res = append(res, packer.CoreBuildPostProcessor{
			PostProcessor:     postProcessor,
			PName:             ppb.PName,
			PType:             ppb.PType,
			KeepInputArtifact: ppb.KeepInputArtifact,
		})
	}

	return res, diags
}

// GetBuilds returns a list of packer Build based on the HCL2 parsed build
// blocks. All Builders, Provisioners and Post Processors will be started and
// configured.
func (cfg *PackerConfig) GetBuilds(opts packer.GetBuildsOptions) ([]packer.Build, hcl.Diagnostics) {
	res := []packer.Build{}
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for _, from := range build.Sources {
			src, found := cfg.Sources[from.Ref()]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + sourceLabel + " " + from.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   fmt.Sprintf("Known: %v", cfg.Sources),
				})
				continue
			}
			src.addition = from.addition
			src.LocalName = from.LocalName

			pcb := &packer.CoreBuild{
				BuildName: build.Name,
				Type:      src.String(),
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

			builder, moreDiags, generatedVars := cfg.startBuilder(src, cfg.EvalContext(nil), opts)
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

			variables := map[string]cty.Value{
				sourcesAccessor: cty.ObjectVal(src.ctyValues()),
				buildAccessor:   cty.ObjectVal(unknownBuildValues),
			}

			provisioners, moreDiags := cfg.getCoreBuildProvisioners(src, build.ProvisionerBlocks, cfg.EvalContext(variables))
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			postProcessors, moreDiags := cfg.getCoreBuildPostProcessors(src, build.PostProcessors, cfg.EvalContext(variables))
			pps := [][]packer.CoreBuildPostProcessor{}
			if len(postProcessors) > 0 {
				pps = [][]packer.CoreBuildPostProcessor{postProcessors}
			} // TODO(azr): remove this
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
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
	for _, v := range p.InputVariables {
		val, _ := v.Value()
		fmt.Fprintf(out, "var.%s: %q [debug: %#v]\n", v.Name, PrintableCtyValue(val), v)
	}
	out.WriteString("\n> local-variables:\n\n")
	for _, v := range p.LocalVariables {
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
			fmt.Fprintf(out, "\n      %s\n", source)
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
		fmt.Fprintf(out, "\n    post-processors:\n\n")
		if len(build.PostProcessors) == 0 {
			fmt.Fprintf(out, "      <no post-processor>\n")
		}
		for _, pp := range build.PostProcessors {
			str := pp.PType
			if pp.PName != "" {
				str = strings.Join([]string{pp.PType, pp.PName}, ".")
			}
			fmt.Fprintf(out, "      %s\n", str)
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

	val, valueDiags := expr.Value(p.EvalContext(nil))
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
