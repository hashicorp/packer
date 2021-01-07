package hcl2template

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// SourceBlock references an HCL 'source' block.
type SourceBlock struct {
	// Type of source; ex: virtualbox-iso
	Type string
	// Given name; if any
	Name string

	block *hcl.Block

	// addition will be merged into block to allow user to override builder settings
	// per build.source block.
	addition hcl.Body
	// LocalName can be set in a singular source block from a build block, it
	// allows to give a special name to a build in the logs.
	LocalName string
}

func (b *SourceBlock) name() string {
	if b.LocalName != "" {
		return b.LocalName
	}
	return b.Name
}

func (b *SourceBlock) String() string {
	return fmt.Sprintf("%s.%s", b.Type, b.name())
}

// EvalContext adds the values of the source to the passed eval context.
func (b *SourceBlock) ctyValues() map[string]cty.Value {
	return map[string]cty.Value{
		"type": cty.StringVal(b.Type),
		"name": cty.StringVal(b.name()),
	}
}

// decodeBuildSource reads a used source block from a build:
//  build {
//    source "type.example" {
//      name = "local_name"
//    }
//  }
func (p *Parser) decodeBuildSource(block *hcl.Block) (SourceRef, hcl.Diagnostics) {
	ref := sourceRefFromString(block.Labels[0])
	var b struct {
		Name string   `hcl:"name,optional"`
		Rest hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return ref, diags
	}
	ref.addition = b.Rest
	ref.LocalName = b.Name
	return ref, nil
}

func (p *Parser) decodeSource(block *hcl.Block) (SourceBlock, hcl.Diagnostics) {
	source := SourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}
	var diags hcl.Diagnostics

	if !p.BuilderSchemas.Has(source.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + buildSourceLabel + " type " + source.Type,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known builders: %v", p.BuilderSchemas.List()),
			Severity: hcl.DiagError,
		})
		return source, diags
	}

	return source, diags
}

func (cfg *PackerConfig) startBuilder(source SourceBlock, ectx *hcl.EvalContext, opts packer.GetBuildsOptions) (packersdk.Builder, hcl.Diagnostics, []string) {
	var diags hcl.Diagnostics

	builder, err := cfg.builderSchemas.Start(source.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed to load " + sourceLabel + " type",
			Detail:  err.Error(),
			Subject: &source.block.LabelRanges[0],
		})
		return builder, diags, nil
	}

	body := source.block.Body
	if source.addition != nil {
		body = hcl.MergeBodies([]hcl.Body{source.block.Body, source.addition})
	}

	decoded, moreDiags := decodeHCL2Spec(body, ectx, builder)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags, nil
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
	builderVars := source.builderVariables()
	builderVars["packer_debug"] = strconv.FormatBool(opts.Debug)
	builderVars["packer_force"] = strconv.FormatBool(opts.Force)
	builderVars["packer_on_error"] = opts.OnError

	generatedVars, warning, err := builder.Prepare(builderVars, decoded)
	moreDiags = warningErrorsToDiags(source.block, warning, err)
	diags = append(diags, moreDiags...)
	return builder, diags, generatedVars
}

// These variables will populate the PackerConfig inside of the builders.
func (source *SourceBlock) builderVariables() map[string]string {
	return map[string]string{
		"packer_build_name":   source.Name,
		"packer_builder_type": source.Type,
	}
}

func (source *SourceBlock) Ref() SourceRef {
	return SourceRef{
		Type: source.Type,
		Name: source.Name,
	}
}

type SourceRef struct {
	Type string
	Name string

	// The content of this body will be merged into a new block to allow to
	// override builder settings per build section.
	addition hcl.Body
	// LocalName can be set in a singular source block from a build block, it
	// allows to give a special name to a build in the logs.
	LocalName string
}

// the 'addition' field makes of ref a different entry in the sources map, so
// Ref is here to make sure only one is returned.
func (r *SourceRef) Ref() SourceRef {
	return SourceRef{
		Type: r.Type,
		Name: r.Name,
	}
}

// NoSource is the zero value of sourceRef, representing the absense of an
// source.
var NoSource SourceRef

func (r SourceRef) String() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}
