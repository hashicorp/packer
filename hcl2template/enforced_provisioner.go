// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/internal/enforcedparser"
	"github.com/hashicorp/packer/packer"
)

func provisionerBlockFromEnforced(pb *enforcedparser.ProvisionerBlock) *ProvisionerBlock {
	return &ProvisionerBlock{
		PType:       pb.PType,
		PName:       pb.PName,
		PauseBefore: pb.PauseBefore,
		MaxRetries:  pb.MaxRetries,
		Timeout:     pb.Timeout,
		Override:    pb.Override,
		OnlyExcept: OnlyExcept{
			Only:   pb.OnlyExcept.Only,
			Except: pb.OnlyExcept.Except,
		},
		HCL2Ref: HCL2Ref{
			DefRange:     pb.DefRange,
			TypeRange:    pb.TypeRange,
			LabelsRanges: pb.LabelsRange,
			Rest:         pb.Rest,
		},
	}
}

// GetCoreBuildProvisionerFromEnforcedBlock converts a shared enforced provisioner block
// into a CoreBuildProvisioner using HCL runtime semantics.
func (cfg *PackerConfig) GetCoreBuildProvisionerFromEnforcedBlock(pb *enforcedparser.ProvisionerBlock, buildName string) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
	return cfg.GetCoreBuildProvisionerFromBlock(provisionerBlockFromEnforced(pb), buildName)
}

// GetCoreBuildProvisionerFromBlock converts a ProvisionerBlock to a CoreBuildProvisioner.
// This is used for enforced provisioners that need to be injected into builds.
func (cfg *PackerConfig) GetCoreBuildProvisionerFromBlock(pb *ProvisionerBlock, buildName string) (packer.CoreBuildProvisioner, hcl.Diagnostics) {
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
		"packer_sensitive_variables": cfg.sensitiveInputVariableKeys(),
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

	if pb.Override != nil {
		if override, ok := pb.Override[buildName]; ok {
			if typedOverride, ok := override.(map[string]interface{}); ok {
				hclProvisioner.override = typedOverride
			}
		}
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
