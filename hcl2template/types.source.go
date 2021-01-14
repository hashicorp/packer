package hcl2template

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// SourceBlock references an HCL 'source' block to be used in a build for
// example.
type SourceBlock struct {
	// Type of source; ex: virtualbox-iso
	Type string
	// Given name; if any
	Name string

	block *hcl.Block

	// LocalName can be set in a singular source block from a build block, it
	// allows to give a special name to a build in the logs.
	LocalName string
}

// SourceBlock references an HCL 'source' block usage.
type SourceUseBlock struct {
	// reference to an actual source block definition, or SourceBlock.
	SourceRef

	// Rest of the body, in case the build.source block has more specific
	// content
	Body hcl.Body
}

func (b *SourceUseBlock) name() string {
	if b.LocalName != "" {
		return b.LocalName
	}
	return b.Name
}

func (b *SourceUseBlock) String() string {
	return fmt.Sprintf("%s.%s", b.Type, b.name())
}

// EvalContext adds the values of the source to the passed eval context.
func (b *SourceUseBlock) ctyValues() map[string]cty.Value {
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
func (p *Parser) decodeBuildSource(block *hcl.Block) (SourceUseBlock, hcl.Diagnostics) {
	ref := sourceRefFromString(block.Labels[0])
	out := SourceUseBlock{SourceRef: ref}
	var b struct {
		Name string   `hcl:"name,optional"`
		Rest hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return out, diags
	}
	out.SourceRef.LocalName = b.Name
	out.Body = b.Rest
	return out, nil
}

func (p *Parser) decodeSource(block *hcl.Block) (SourceBlock, hcl.Diagnostics) {
	source := SourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}
	var diags hcl.Diagnostics

	return source, diags
}

func (cfg *PackerConfig) startBuilder(source SourceUseBlock, ectx *hcl.EvalContext, opts packer.GetBuildsOptions) (packersdk.Builder, hcl.Diagnostics, []string) {
	var diags hcl.Diagnostics

	builder, err := cfg.parser.PluginConfig.Builders.Start(source.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed to load " + sourceLabel + " type",
			Detail:  err.Error(),
		})
		return builder, diags, nil
	}

	body := source.Body
	if body == nil {
		panic("body is nil")
	}
	decoded, moreDiags := decodeHCL2Spec(body, ectx, builder)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, diags, nil
	}

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
	moreDiags = warningErrorsToDiags(cfg.Sources[source.SourceRef.Ref()].block, warning, err)
	diags = append(diags, moreDiags...)
	return builder, diags, generatedVars
}

// These variables will populate the PackerConfig inside of the builders.
func (source *SourceUseBlock) builderVariables() map[string]string {
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

// SourceRef is a nice way to put `virtualbox-iso.source_name`
type SourceRef struct {
	// Type of the source, for example `virtualbox-iso`
	Type string
	// Name of the source, for example `source_name`
	Name string

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
