// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// ProvisionerBlock references a detected but unparsed post processor
type PostProcessorBlock struct {
	PType             string
	PName             string
	OnlyExcept        OnlyExcept
	KeepInputArtifact *bool

	HCL2Ref
}

func (p *PostProcessorBlock) String() string {
	return fmt.Sprintf(BuildPostProcessorLabel+"-block %q %q", p.PType, p.PName)
}

func (p *Parser) decodePostProcessor(block *hcl.Block, ectx *hcl.EvalContext) (*PostProcessorBlock, hcl.Diagnostics) {
	var b struct {
		Name              string   `hcl:"name,optional"`
		Only              []string `hcl:"only,optional"`
		Except            []string `hcl:"except,optional"`
		KeepInputArtifact *bool    `hcl:"keep_input_artifact,optional"`
		Rest              hcl.Body `hcl:",remain"`
	}

	diags := gohcl.DecodeBody(block.Body, ectx, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	postProcessor := &PostProcessorBlock{
		PType:             block.Labels[0],
		PName:             b.Name,
		OnlyExcept:        OnlyExcept{Only: b.Only, Except: b.Except},
		HCL2Ref:           newHCL2Ref(block, b.Rest),
		KeepInputArtifact: b.KeepInputArtifact,
	}

	diags = diags.Extend(postProcessor.OnlyExcept.Validate())
	if diags.HasErrors() {
		return nil, diags
	}

	return postProcessor, diags
}
