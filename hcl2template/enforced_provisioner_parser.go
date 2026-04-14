// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

var standaloneProvisionerSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildProvisionerLabel, LabelNames: []string{"type"}},
	},
}

// ParseProvisionerBlocks parses raw provisioner block content into ProvisionerBlocks.
// It accepts HCL, HCL JSON, and the legacy JSON payload used for enforced provisioners.
func ParseProvisionerBlocks(blockContent string) ([]*ProvisionerBlock, hcl.Diagnostics) {
	parser := &Parser{Parser: hclparse.NewParser()}
	return parser.parseProvisionerBlocks(blockContent)
}

func (p *Parser) parseProvisionerBlocks(blockContent string) ([]*ProvisionerBlock, hcl.Diagnostics) {
	hclParser := p.Parser
	if hclParser == nil {
		hclParser = hclparse.NewParser()
	}

	log.Printf("[DEBUG] parsing provisioner block content as HCL")

	file, diags := hclParser.ParseHCL([]byte(blockContent), "provisioner.pkr.hcl")
	if !diags.HasErrors() {
		log.Printf("[DEBUG] parsed provisioner block content as HCL")
		return p.parseProvisionerBlocksFromFile(file, diags)
	}
	log.Printf("[DEBUG] failed to parse provisioner block content as HCL, trying JSON fallback")

	jsonFile, jsonDiags := hclParser.ParseJSON([]byte(blockContent), "provisioner.pkr.json")
	if jsonDiags.HasErrors() {
		log.Printf("[DEBUG] failed to parse provisioner block content as JSON")
		return nil, append(diags, jsonDiags...)
	}

	provisioners, provisionerDiags := p.parseProvisionerBlocksFromFile(jsonFile, jsonDiags)
	if !provisionerDiags.HasErrors() && len(provisioners) > 0 {
		log.Printf("[DEBUG] parsed provisioner block content as JSON")
		return provisioners, provisionerDiags
	}

	legacyJSON, ok, err := normalizeLegacyProvisionersJSON(blockContent)
	if err == nil && ok {
		legacyFile, legacyDiags := hclParser.ParseJSON([]byte(legacyJSON), "provisioner_legacy.pkr.json")
		if !legacyDiags.HasErrors() {
			legacyProvisioners, legacyProvisionerDiags := p.parseProvisionerBlocksFromFile(legacyFile, legacyDiags)
			if !legacyProvisionerDiags.HasErrors() && len(legacyProvisioners) > 0 {
				log.Printf("[DEBUG] parsed provisioner block content as legacy JSON")
				return legacyProvisioners, legacyProvisionerDiags
			}
		}
	}

	if provisionerDiags.HasErrors() {
		return nil, provisionerDiags
	}
	log.Printf("[DEBUG] parsed provisioner block content as JSON but found no valid provisioner blocks")
	return provisioners, provisionerDiags
}

func normalizeLegacyProvisionersJSON(blockContent string) (string, bool, error) {
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
	for _, provisioner := range payload.Provisioners {
		typeName, ok := provisioner["type"].(string)
		if !ok || typeName == "" {
			continue
		}

		cfg := make(map[string]interface{})
		for key, value := range provisioner {
			if key == "type" {
				continue
			}
			cfg[key] = value
		}

		normalized = append(normalized, map[string]interface{}{typeName: cfg})
	}

	if len(normalized) == 0 {
		return "", false, nil
	}

	out := map[string]interface{}{
		buildProvisionerLabel: normalized,
	}

	b, err := json.Marshal(out)
	if err != nil {
		return "", false, err
	}

	return string(b), true, nil
}

func (p *Parser) parseProvisionerBlocksFromFile(file *hcl.File, diags hcl.Diagnostics) ([]*ProvisionerBlock, hcl.Diagnostics) {
	content, moreDiags := file.Body.Content(standaloneProvisionerSchema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}

	ectx := &hcl.EvalContext{Variables: map[string]cty.Value{}}
	provisioners := make([]*ProvisionerBlock, 0, len(content.Blocks))

	for _, block := range content.Blocks {
		provisioner, moreDiags := p.decodeProvisioner(block, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		provisioners = append(provisioners, provisioner)
	}

	return provisioners, diags
}
