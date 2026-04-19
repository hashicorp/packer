// Copyright IBM Corp. 2013, 2025
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

	// Body is the source block body with the top-level `tags` and `labels`
	// attributes stripped out. This is what downstream code (body merging,
	// plugin decoding) should use. It is populated at parse time; if the
	// source has no tags/labels it is equal to block.Body.
	Body hcl.Body

	// Tags is the list of tags declared on the source block.
	Tags []string
	// Labels is the key/value metadata declared on the source block.
	Labels map[string]string

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

	// Tags is the list of tags declared on an inline `source "type.name" {}`
	// usage inside a build block. These are layered on top of the definition
	// source's tags.
	Tags []string
	// Labels is the key/value metadata declared on an inline source usage
	// inside a build block. Layered on top of the definition source's labels
	// (usage keys win on conflict with the definition source, but the
	// definition source still wins over the enclosing build block).
	Labels map[string]string

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

// metadataBlockLabel is the reserved nested block name that carries
// Packer-specific build metadata (tags and labels) on source and build
// blocks. It is stripped from the body before the plugin ConfigSpec
// decoder runs, so it never collides with plugin-defined attributes such
// as amazon-ebs's own `tags` map.
const metadataBlockLabel = "metadata"

// metadataBodySchema describes the attributes accepted inside a
// `metadata { }` nested block.
type metadataBody struct {
	Tags   []string          `hcl:"tags,optional"`
	Labels map[string]string `hcl:"labels,optional"`
}

// extractMetadata pulls any `metadata` block out of body. It returns the
// extracted tags/labels, a remainder body with the metadata block removed,
// and any diagnostics. When body has no metadata block, tags and labels
// are nil and remainder == body.
func extractMetadata(body hcl.Body, ectx *hcl.EvalContext) (tags []string, labels map[string]string, remainder hcl.Body, diags hcl.Diagnostics) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: metadataBlockLabel},
		},
	}
	content, remain, diags := body.PartialContent(schema)
	if len(content.Blocks) == 0 {
		// No metadata block: return the original body so callers can detect
		// the no-op case by pointer equality and avoid perturbing downstream
		// test fixtures.
		return nil, nil, body, diags
	}
	remainder = remain
	if len(content.Blocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Only one %q block is allowed per source or build", metadataBlockLabel),
			Subject:  content.Blocks[1].DefRange.Ptr(),
		})
	}
	mb := content.Blocks[0]
	var decoded metadataBody
	moreDiags := gohcl.DecodeBody(mb.Body, ectx, &decoded)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return nil, nil, remainder, diags
	}
	return dedupStrings(decoded.Tags), decoded.Labels, remainder, diags
}

// decodeBuildSource reads a used source block from a build:
//
//	build {
//	  source "type.example" {
//	    name = "local_name"
//	    metadata {
//	      tags   = ["prod"]
//	      labels = { region = "us-east" }
//	    }
//	  }
//	}
func (p *Parser) decodeBuildSource(block *hcl.Block) (SourceUseBlock, hcl.Diagnostics) {
	ref := sourceRefFromString(block.Labels[0])
	out := SourceUseBlock{SourceRef: ref}

	// First strip out the metadata block so the subsequent gohcl decode
	// and the plugin ConfigSpec decoder never see it.
	tags, labels, bodyAfterMeta, diags := extractMetadata(block.Body, nil)
	if diags.HasErrors() {
		return out, diags
	}
	out.Tags = tags
	out.Labels = labels

	var b struct {
		Name string   `hcl:"name,optional"`
		Rest hcl.Body `hcl:",remain"`
	}
	moreDiags := gohcl.DecodeBody(bodyAfterMeta, nil, &b)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return out, diags
	}
	out.LocalName = b.Name
	out.Body = b.Rest
	return out, diags
}

func (p *Parser) decodeSource(block *hcl.Block) (SourceBlock, hcl.Diagnostics) {
	source := SourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}
	tags, labels, remain, diags := extractMetadata(block.Body, nil)
	// Only populate the filter-specific fields when the user actually
	// declared a metadata block. Leaving Body/Tags/Labels as their zero
	// values in the common case avoids perturbing equality checks in
	// existing parser tests, and plugin.go falls back to block.Body when
	// Body is nil.
	if tags != nil || labels != nil {
		source.Body = remain
		source.Tags = tags
		source.Labels = labels
	}
	return source, diags
}

// dedupStrings returns s with duplicates removed, preserving order. Returns
// nil when s is empty so callers can distinguish "no tags declared" from
// "empty tag list declared".
func dedupStrings(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(s))
	out := make([]string, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
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
func (source *SourceUseBlock) builderVariables() map[string]interface{} {
	return map[string]interface{}{
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
