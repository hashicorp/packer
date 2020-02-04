package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// PackerConfig represents a loaded Packer HCL config. It will contain
// references to all possible blocks of the allowed configuration.
type PackerConfig struct {
	// Directory where the config files are defined
	Basedir string

	// Available Source blocks
	Sources map[SourceRef]*SourceBlock

	// InputVariables and LocalVariables are the list of defined input and
	// local variables. They are of the same type but are not used in the same
	// way. Local variables will not be decoded from any config file, env var,
	// or ect. Like the Input variables will.
	InputVariables Variables
	LocalVariables Variables

	// Builds is the list of Build blocks defined in the config files.
	Builds Builds
}

// EvalContext returns the *hcl.EvalContext that will be passed to an hcl
// decoder in order to tell what is the actual value of a var or a local and
// the list of defined functions.
func (cfg *PackerConfig) EvalContext() *hcl.EvalContext {
	ectx := &hcl.EvalContext{
		Functions: Functions(cfg.Basedir),
		Variables: map[string]cty.Value{
			"var":   cty.ObjectVal(cfg.InputVariables.Values()),
			"local": cty.ObjectVal(cfg.LocalVariables.Values()),
		},
	}
	return ectx
}

// getCoreBuildProvisioners takes a list of provisioner block, starts according
// provisioners and sends parsed HCL2 over to it.
func (p *Parser) getCoreBuildProvisioners(blocks []*ProvisionerBlock, ectx *hcl.EvalContext, generatedVars []string) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		provisioner, moreDiags := p.StartProvisioner(pb, ectx, generatedVars)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, packer.CoreBuildProvisioner{
			PType:       pb.PType,
			PName:       pb.PName,
			Provisioner: provisioner,
		})
	}
	return res, diags
}

// getCoreBuildProvisioners takes a list of post processor block, starts according
// provisioners and sends parsed HCL2 over to it.
func (p *Parser) getCoreBuildPostProcessors(blocks []*PostProcessorBlock, ectx *hcl.EvalContext) ([]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildPostProcessor{}
	for _, ppb := range blocks {
		postProcessor, moreDiags := p.StartPostProcessor(ppb, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, packer.CoreBuildPostProcessor{
			PostProcessor: postProcessor,
			PName:         ppb.PName,
			PType:         ppb.PType,
		})
	}

	return res, diags
}

// getBuilds will return a list of packer Build based on the HCL2 parsed build
// blocks. All Builders, Provisioners and Post Processors will be started and
// configured.
func (p *Parser) getBuilds(cfg *PackerConfig) ([]packer.Build, hcl.Diagnostics) {
	res := []packer.Build{}
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for _, from := range build.Sources {
			src, found := cfg.Sources[from]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + sourceLabel + " " + from.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
				})
				continue
			}
			builder, moreDiags, generatedVars := p.StartBuilder(src, cfg.EvalContext())
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			provisioners, moreDiags := p.getCoreBuildProvisioners(build.ProvisionerBlocks, cfg.EvalContext(), generatedVars)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			postProcessors, moreDiags := p.getCoreBuildPostProcessors(build.PostProcessors, cfg.EvalContext())
			pps := [][]packer.CoreBuildPostProcessor{}
			if len(postProcessors) > 0 {
				pps = [][]packer.CoreBuildPostProcessor{postProcessors}
			}
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			pcb := &packer.CoreBuild{
				Type:           src.Type,
				Builder:        builder,
				Provisioners:   provisioners,
				PostProcessors: pps,
			}
			res = append(res, pcb)
		}
	}
	return res, diags
}

// Parse will parse HCL file(s) in path. Path can be a folder or a file.
//
// Parse will first parse variables and then the rest; so that interpolation
// can happen.
//
// Parse then return a slice of packer.Builds; which are what packer core uses
// to run builds.
func (p *Parser) Parse(path string, vars map[string]string) ([]packer.Build, hcl.Diagnostics) {
	cfg, diags := p.parse(path, vars)
	if diags.HasErrors() {
		return nil, diags
	}

	builds, moreDiags := p.getBuilds(cfg)
	return builds, append(diags, moreDiags...)
}
