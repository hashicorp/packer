package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerAmazonEnhancedNetworking is a Fixer that replaces the "enhanced_networking" configuration key
// with the clearer "ena_support".  This disambiguates ena_support from sriov_support.
type FixerAmazonEnhancedNetworking struct{}

func (FixerAmazonEnhancedNetworking) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each builder and replace the enhanced_networking if we can
	for _, builder := range tpl.Builders {
		enhancedNetworkingRaw, ok := builder["enhanced_networking"]
		if !ok {
			continue
		}
		enhancedNetworkingString, ok := enhancedNetworkingRaw.(bool)
		if !ok {
			// TODO: error?
			continue
		}

		delete(builder, "enhanced_networking")
		builder["ena_support"] = enhancedNetworkingString
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAmazonEnhancedNetworking) Synopsis() string {
	return `Replaces "enhanced_networking" in builders with "ena_support"`
}
