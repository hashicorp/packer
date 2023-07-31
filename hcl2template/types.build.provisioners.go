// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/zclconf/go-cty/cty"
)

// OnlyExcept is a struct that is meant to be embedded that contains the
// logic required for "only" and "except" meta-parameters.
type OnlyExcept struct {
	Only   []string `json:"only,omitempty"`
	Except []string `json:"except,omitempty"`
}

// Skip says whether or not to skip the build with the given name.
func (o *OnlyExcept) Skip(n string) bool {
	if len(o.Only) > 0 {
		for _, v := range o.Only {
			if v == n {
				return false
			}
		}

		return true
	}

	if len(o.Except) > 0 {
		for _, v := range o.Except {
			if v == n {
				return true
			}
		}

		return false
	}

	return false
}

// Validate validates that the OnlyExcept settings are correct for a thing.
func (o *OnlyExcept) Validate() hcl.Diagnostics {
	var diags hcl.Diagnostics

	if len(o.Only) > 0 && len(o.Except) > 0 {
		diags = diags.Append(&hcl.Diagnostic{
			Summary:  "only one of 'only' or 'except' may be specified",
			Severity: hcl.DiagError,
		})
	}

	return diags
}

// ProvisionerBlock references a detected but unparsed provisioner
type ProvisionerBlock struct {
	PType       string
	PName       string
	PauseBefore time.Duration
	MaxRetries  int
	Timeout     time.Duration
	Override    map[string]interface{}
	OnlyExcept  OnlyExcept
	HCL2Ref
}

func (p *ProvisionerBlock) String() string {
	return fmt.Sprintf(BuildProvisionerLabel+"-block %q %q", p.PType, p.PName)
}

func (p *Parser) decodeProvisioner(block *hcl.Block, ectx *hcl.EvalContext) (*ProvisionerBlock, hcl.Diagnostics) {
	var b struct {
		Name        string    `hcl:"name,optional"`
		PauseBefore string    `hcl:"pause_before,optional"`
		MaxRetries  int       `hcl:"max_retries,optional"`
		Timeout     string    `hcl:"timeout,optional"`
		Only        []string  `hcl:"only,optional"`
		Except      []string  `hcl:"except,optional"`
		Override    cty.Value `hcl:"override,optional"`
		Rest        hcl.Body  `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, ectx, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	provisioner := &ProvisionerBlock{
		PType:      block.Labels[0],
		PName:      b.Name,
		MaxRetries: b.MaxRetries,
		OnlyExcept: OnlyExcept{Only: b.Only, Except: b.Except},
		HCL2Ref:    newHCL2Ref(block, b.Rest),
	}

	diags = diags.Extend(provisioner.OnlyExcept.Validate())
	if diags.HasErrors() {
		return nil, diags
	}

	if !b.Override.IsNull() {
		if !b.Override.Type().IsObjectType() {
			return nil, append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "provisioner's override block must be an HCL object",
				Subject:  block.DefRange.Ptr(),
			})
		}

		override := make(map[string]interface{})
		for buildName, overrides := range b.Override.AsValueMap() {
			buildOverrides := make(map[string]interface{})

			if !overrides.Type().IsObjectType() {
				return nil, append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: fmt.Sprintf(
						"provisioner's override.'%s' block must be an HCL object",
						buildName),
					Subject: block.DefRange.Ptr(),
				})
			}

			for option, value := range overrides.AsValueMap() {
				buildOverrides[option] = hcl2shim.ConfigValueFromHCL2(value)
			}
			override[buildName] = buildOverrides
		}
		provisioner.Override = override
	}

	if b.PauseBefore != "" {
		pauseBefore, err := time.ParseDuration(b.PauseBefore)
		if err != nil {
			return nil, append(diags, &hcl.Diagnostic{
				Summary:  "Failed to parse pause_before duration",
				Severity: hcl.DiagError,
				Detail:   err.Error(),
				Subject:  &block.DefRange,
			})
		}
		provisioner.PauseBefore = pauseBefore
	}

	if b.Timeout != "" {
		timeout, err := time.ParseDuration(b.Timeout)
		if err != nil {
			return nil, append(diags, &hcl.Diagnostic{
				Summary:  "Failed to parse timeout duration",
				Severity: hcl.DiagError,
				Detail:   err.Error(),
				Subject:  &block.DefRange,
			})
		}
		provisioner.Timeout = timeout
	}

	return provisioner, diags
}
