// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import "github.com/mitchellh/mapstructure"

// FixerVagrantPPOverride is a Fixer that replaces the provider-specific
// overrides for the Vagrant post-processor with the new style introduced
// as part of Packer 0.5.0.
type FixerVagrantPPOverride struct{}

func (FixerVagrantPPOverride) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerVagrantPPOverride) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	if input["post-processors"] == nil {
		return input, nil
	}

	// Our template type we'll use for this fixer only
	type template struct {
		PP `mapstructure:",squash"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	pps := tpl.ppList()

	// Go through each post-processor and make the fix if necessary
	possible := []string{"aws", "digitalocean", "virtualbox", "vmware"}
	for _, pp := range pps {
		typeRaw, ok := pp["type"]
		if !ok {
			continue
		}

		if typeName, ok := typeRaw.(string); !ok {
			continue
		} else if typeName != "vagrant" {
			continue
		}

		overrides := make(map[string]interface{})
		for _, name := range possible {
			if _, ok := pp[name]; !ok {
				continue
			}

			overrides[name] = pp[name]
			delete(pp, name)
		}

		if len(overrides) > 0 {
			pp["override"] = overrides
		}
	}

	input["post-processors"] = tpl.PostProcessors
	return input, nil
}

func (FixerVagrantPPOverride) Synopsis() string {
	return `Fixes provider-specific overrides for Vagrant post-processor`
}
