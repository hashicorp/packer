package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVagrantPPOvveride is a Fixer that replaces the provider-specific
// overrides for the Vagrant post-processor with the new style introduced
// as part of Packer 0.5.0.
type FixerVagrantPPOverride struct{}

func (FixerVagrantPPOverride) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		PostProcessors []interface{} `mapstructure:"post-processors"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each post-processor and get out all the complex configs
	pps := make([]map[string]interface{}, 0, len(tpl.PostProcessors))
	for _, rawPP := range tpl.PostProcessors {
		switch pp := rawPP.(type) {
		case string:
		case map[string]interface{}:
			pps = append(pps, pp)
		case []interface{}:
			for _, innerRawPP := range pp {
				if innerPP, ok := innerRawPP.(map[string]interface{}); ok {
					pps = append(pps, innerPP)
				}
			}
		}
	}

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
