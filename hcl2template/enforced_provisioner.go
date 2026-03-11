// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

var enforcedProvisionerSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildProvisionerLabel, LabelNames: []string{"type"}},
	},
}

// ParseProvisionerBlocks parses a partial HCL string that contains only
// top-level provisioner blocks and returns the parsed ProvisionerBlock list.
func ParseProvisionerBlocks(blockContent string) ([]*ProvisionerBlock, hcl.Diagnostics) {
	parser := &Parser{Parser: hclparse.NewParser()}
	file, diags := parser.ParseHCL([]byte(blockContent), "enforced_provisioner.pkr.hcl")
	if diags.HasErrors() {
		return nil, diags
	}

	content, moreDiags := file.Body.Content(enforcedProvisionerSchema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}

	ectx := &hcl.EvalContext{Variables: map[string]cty.Value{}}
	provisioners := make([]*ProvisionerBlock, 0, len(content.Blocks))

	for _, block := range content.Blocks {
		prov, moreDiags := parser.decodeProvisioner(block, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		provisioners = append(provisioners, prov)
	}

	return provisioners, diags
}

// GetCoreBuildProvisionerFromBlock converts a ProvisionerBlock to a CoreBuildProvisioner.
// This is used for enforced provisioners that need to be injected into builds.
func (cfg *PackerConfig) GetCoreBuildProvisionerFromBlock(pb *ProvisionerBlock) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// Get the provisioner plugin
	provisioner, err := cfg.parser.PluginConfig.Provisioners.Start(pb.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to start enforced provisioner %q", pb.PType),
			Detail:   fmt.Sprintf("The provisioner plugin could not be loaded: %s", err.Error()),
		})
		return packer.CoreBuildProvisioner{}, diags
	}

	// Create basic builder variables
	builderVars := map[string]interface{}{
		"packer_core_version":        cfg.CorePackerVersionString,
		"packer_debug":               strconv.FormatBool(cfg.debug),
		"packer_force":               strconv.FormatBool(cfg.force),
		"packer_on_error":            cfg.onError,
		"packer_sensitive_variables": []string{},
	}

	// Create evaluation context
	ectx := cfg.EvalContext(BuildContext, nil)

	// Create the HCL2Provisioner wrapper
	hclProvisioner := &HCL2Provisioner{
		Provisioner:      provisioner,
		provisionerBlock: pb,
		evalContext:      ectx,
		builderVariables: builderVars,
	}

	// Prepare the provisioner
	err = hclProvisioner.HCL2Prepare(nil)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to prepare enforced provisioner %q", pb.PType),
			Detail:   err.Error(),
		})
		return packer.CoreBuildProvisioner{}, diags
	}

	// Wrap provisioner with any special behavior (pause, timeout, retry)
	wrappedProvisioner := packer.WrapProvisionerWithOptions(hclProvisioner, packer.ProvisionerWrapOptions{
		PauseBefore: pb.PauseBefore,
		Timeout:     pb.Timeout,
		MaxRetries:  pb.MaxRetries,
	})

	return packer.CoreBuildProvisioner{
		PType:       pb.PType,
		PName:       pb.PName,
		Provisioner: wrappedProvisioner,
	}, diags
}
