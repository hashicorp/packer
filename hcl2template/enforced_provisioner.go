// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"encoding/json"
	"fmt"
	"log"
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

// ParseProvisionerBlocks parses a string containing one or more top-level provisioner blocks
// in either HCL or JSON syntax, and returns a slice of parsed ProvisionerBlock objects along
// with any diagnostics encountered during parsing.
func ParseProvisionerBlocks(blockContent string) ([]*ProvisionerBlock, hcl.Diagnostics) {
	parser := &Parser{Parser: hclparse.NewParser()}
	log.Printf("[DEBUG] parsing enforced provisioner block content as HCL")

	file, diags := parser.ParseHCL([]byte(blockContent), "enforced_provisioner.pkr.hcl")
	if !diags.HasErrors() {
		log.Printf("[DEBUG] parsed enforced provisioner block content as HCL")
		return parseProvisionerBlocksFromFile(parser, file, diags)
	}
	log.Printf("[DEBUG] failed to parse enforced provisioner block content as HCL, trying JSON fallback")

	// Fallback to HCL-JSON for enforced block content authored in JSON syntax.
	jsonFile, jsonDiags := parser.ParseJSON([]byte(blockContent), "enforced_provisioner.pkr.json")
	if jsonDiags.HasErrors() {
		log.Printf("[DEBUG] failed to parse enforced provisioner block content as JSON")
		return nil, append(diags, jsonDiags...)
	}

	provisioners, provisionerDiags := parseProvisionerBlocksFromFile(parser, jsonFile, jsonDiags)
	if !provisionerDiags.HasErrors() && len(provisioners) > 0 {
		log.Printf("[DEBUG] parsed enforced provisioner block content as JSON")
		return provisioners, provisionerDiags
	}

	// Backward compatibility fallback for legacy JSON shape:
	// {"provisioners":[{"type":"shell", ...}]}
	legacyJSON, ok, err := normalizeLegacyEnforcedProvisionersJSON(blockContent)
	if err == nil && ok {
		legacyFile, legacyDiags := parser.ParseJSON([]byte(legacyJSON), "enforced_provisioner_legacy.pkr.json")
		if !legacyDiags.HasErrors() {
			legacyProvisioners, legacyProvisionerDiags := parseProvisionerBlocksFromFile(parser, legacyFile, legacyDiags)
			if !legacyProvisionerDiags.HasErrors() && len(legacyProvisioners) > 0 {
				log.Printf("[DEBUG] parsed enforced provisioner block content as legacy JSON")
				return legacyProvisioners, legacyProvisionerDiags
			}
		}
	}

	if provisionerDiags.HasErrors() {
		return nil, provisionerDiags
	}
	log.Printf("[DEBUG] parsed enforced provisioner block content as JSON but found no valid provisioner blocks")
	return provisioners, provisionerDiags
}

func normalizeLegacyEnforcedProvisionersJSON(blockContent string) (string, bool, error) {
	type legacyPayload struct {
		Provisioners []map[string]interface{} `json:"provisioners"`
	}

	var payload legacyPayload
	if err := json.Unmarshal([]byte(blockContent), &payload); err != nil {
		return "", false, err
	}

	if len(payload.Provisioners) == 0 {
		return "", false, nil
	}

	normalized := make([]map[string]interface{}, 0, len(payload.Provisioners))
	for _, p := range payload.Provisioners {
		typeName, ok := p["type"].(string)
		if !ok || typeName == "" {
			continue
		}

		cfg := make(map[string]interface{})
		for k, v := range p {
			if k == "type" {
				continue
			}
			cfg[k] = v
		}

		normalized = append(normalized, map[string]interface{}{typeName: cfg})
	}

	if len(normalized) == 0 {
		return "", false, nil
	}

	out := map[string]interface{}{
		"provisioner": normalized,
	}

	b, err := json.Marshal(out)
	if err != nil {
		return "", false, err
	}

	return string(b), true, nil
}

func parseProvisionerBlocksFromFile(parser *Parser, file *hcl.File, diags hcl.Diagnostics) ([]*ProvisionerBlock, hcl.Diagnostics) {

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
