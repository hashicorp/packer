// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
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

// SourceUseBlock is a SourceBlock 'usage' from a config stand point.
// For example when one uses `build.sources = ["..."]` or
// `build.source "..." {...}`.
type SourceUseBlock struct {
	// reference to an actual source block definition, or SourceBlock.
	SourceRef

	// LocalName can be set in a singular source block from a build block, it
	// allows to give a special name to a build in the logs.
	LocalName string

	// Rest of the body, in case the build.source block has more specific
	// content
	// Body can be expanded by a dynamic tag.
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
//
//	build {
//	  source "type.example" {
//	    name = "local_name"
//	  }
//	}
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
	out.LocalName = b.Name
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

func (cfg *PackerConfig) startBuilder(source SourceUseBlock, ectx *hcl.EvalContext) (packersdk.Builder, hcl.Diagnostics, []string) {
	var diags hcl.Diagnostics

	builder, err := cfg.parser.PluginConfig.Builders.Start(source.Type)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to load " + sourceLabel + " type",
			Detail:   err.Error(),
		})
		return builder, diags, nil
	}

	body := source.Body
	// Add known values to source accessor in eval context.
	ectx.Variables[sourcesAccessor] = cty.ObjectVal(source.ctyValues())

	decoded, moreDiags := decodeHCL2Spec(body, ectx, builder)
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
	builderVars := source.builderVariables()
	builderVars["packer_core_version"] = cfg.CorePackerVersionString
	builderVars["packer_debug"] = strconv.FormatBool(cfg.debug)
	builderVars["packer_force"] = strconv.FormatBool(cfg.force)
	builderVars["packer_on_error"] = cfg.onError

	generatedVars, warning, err := builder.Prepare(builderVars, decoded)
	moreDiags = warningErrorsToDiags(cfg.Sources[source.SourceRef].block, warning, err)
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

	// No other field should be added to the SourceRef because we used that
	// struct as a map accessor in many places.
}

// NoSource is the zero value of sourceRef, representing the absense of an
// source.
var NoSource SourceRef

func (r SourceRef) String() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}

func listAvailableSourceNames(srcs map[SourceRef]SourceBlock) []string {
	res := make([]string, 0, len(srcs))
	for k := range srcs {
		res = append(res, k.String())
	}
	sort.Strings(res)
	return res
}
