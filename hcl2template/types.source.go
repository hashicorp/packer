// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// SourceBlock references an HCL 'source' block to be used in a build for
// example.
type SourceBlock struct {
	// Type of source; ex: virtualbox-iso
	Type string
	// Given name; if any
	SourceName string

	Block *hcl.Block

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

func (b *SourceUseBlock) Name() string {
	if b.LocalName != "" {
		return b.LocalName
	}
	return b.SourceName
}

func (b *SourceUseBlock) String() string {
	return fmt.Sprintf("%s.%s", b.Type, b.Name())
}

// EvalContext adds the values of the source to the passed eval context.
func (b *SourceUseBlock) CtyValues() map[string]cty.Value {
	return map[string]cty.Value{
		"type": cty.StringVal(b.Type),
		"name": cty.StringVal(b.Name()),
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
		Type:       block.Labels[0],
		SourceName: block.Labels[1],
		Block:      block,
	}
	var diags hcl.Diagnostics

	return source, diags
}

// These variables will populate the PackerConfig inside of the builders.
func (source *SourceUseBlock) BuilderVariables() map[string]string {
	return map[string]string{
		"packer_build_name":   source.SourceName,
		"packer_builder_type": source.Type,
	}
}

func (source *SourceBlock) Ref() SourceRef {
	return SourceRef{
		Type:       source.Type,
		SourceName: source.SourceName,
	}
}

// SourceRef is a nice way to put `virtualbox-iso.source_name`
type SourceRef struct {
	// Type of the source, for example `virtualbox-iso`
	Type string
	// Name of the source, for example `source_name`
	SourceName string

	// No other field should be added to the SourceRef because we used that
	// struct as a map accessor in many places.
}

// NoSource is the zero value of sourceRef, representing the absense of an
// source.
var NoSource SourceRef

func (r SourceRef) String() string {
	return fmt.Sprintf("%s.%s", r.Type, r.SourceName)
}

func ListAvailableSourceNames(srcs map[SourceRef]SourceBlock) []string {
	res := make([]string, 0, len(srcs))
	for k := range srcs {
		res = append(res, k.String())
	}
	sort.Strings(res)
	return res
}
