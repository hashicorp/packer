// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package enforcedparser

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/zclconf/go-cty/cty"
)

const provisionerBlockLabel = "provisioner"

var enforcedProvisionerSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: provisionerBlockLabel, LabelNames: []string{"type"}},
	},
}

type OnlyExcept struct {
	Only   []string `json:"only,omitempty"`
	Except []string `json:"except,omitempty"`
}

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

type ProvisionerBlock struct {
	PType       string
	PName       string
	PauseBefore time.Duration
	MaxRetries  int
	Timeout     time.Duration
	Override    map[string]interface{}
	OnlyExcept  OnlyExcept
	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelsRange []hcl.Range
	Rest        hcl.Body
}

// ParseProvisionerBlocks parses raw enforced block content into a neutral provisioner model.
func ParseProvisionerBlocks(blockContent string) ([]*ProvisionerBlock, hcl.Diagnostics) {
	parser := hclparse.NewParser()
	log.Printf("[DEBUG] parsing enforced provisioner block content as HCL")

	file, diags := parser.ParseHCL([]byte(blockContent), "enforced_provisioner.pkr.hcl")
	if !diags.HasErrors() {
		log.Printf("[DEBUG] parsed enforced provisioner block content as HCL")
		return parseProvisionerBlocksFromFile(file, diags)
	}
	log.Printf("[DEBUG] failed to parse enforced provisioner block content as HCL, trying JSON fallback")

	jsonFile, jsonDiags := parser.ParseJSON([]byte(blockContent), "enforced_provisioner.pkr.json")
	if jsonDiags.HasErrors() {
		log.Printf("[DEBUG] failed to parse enforced provisioner block content as JSON")
		return nil, append(diags, jsonDiags...)
	}

	provisioners, provisionerDiags := parseProvisionerBlocksFromFile(jsonFile, jsonDiags)
	if !provisionerDiags.HasErrors() && len(provisioners) > 0 {
		log.Printf("[DEBUG] parsed enforced provisioner block content as JSON")
		return provisioners, provisionerDiags
	}

	legacyJSON, ok, err := normalizeLegacyEnforcedProvisionersJSON(blockContent)
	if err == nil && ok {
		legacyFile, legacyDiags := parser.ParseJSON([]byte(legacyJSON), "enforced_provisioner_legacy.pkr.json")
		if !legacyDiags.HasErrors() {
			legacyProvisioners, legacyProvisionerDiags := parseProvisionerBlocksFromFile(legacyFile, legacyDiags)
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

func parseProvisionerBlocksFromFile(file *hcl.File, diags hcl.Diagnostics) ([]*ProvisionerBlock, hcl.Diagnostics) {
	content, moreDiags := file.Body.Content(enforcedProvisionerSchema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}

	ectx := &hcl.EvalContext{Variables: map[string]cty.Value{}}
	provisioners := make([]*ProvisionerBlock, 0, len(content.Blocks))

	for _, block := range content.Blocks {
		prov, moreDiags := decodeProvisioner(block, ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		provisioners = append(provisioners, prov)
	}

	return provisioners, diags
}

func decodeProvisioner(block *hcl.Block, ectx *hcl.EvalContext) (*ProvisionerBlock, hcl.Diagnostics) {
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
		DefRange:   block.DefRange,
		TypeRange:  block.TypeRange,
		LabelsRange: block.LabelRanges,
		Rest:       b.Rest,
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