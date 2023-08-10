// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerAmazonSpotPriceProductDeprecation removes the deprecated "spot_price_auto_product" setting
// from Amazon builder templates
type FixerAmazonSpotPriceProductDeprecation struct{}

func (FixerAmazonSpotPriceProductDeprecation) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*amazon*": []string{"spot_price_auto_product"},
	}
}

func (FixerAmazonSpotPriceProductDeprecation) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// The type we'll decode into; we only care about builders
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	for _, builder := range tpl.Builders {
		builderTypeRaw, ok := builder["type"]
		if !ok {
			continue
		}

		builderType, ok := builderTypeRaw.(string)
		if !ok {
			continue
		}

		buildersToFix := []string{"amazon-ebs", "amazon-ebssurrogate",
			"amazon-ebsvolume", "amazon-instance"}

		matched := false
		for _, b := range buildersToFix {
			if builderType == b {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		_, ok = builder["spot_price_auto_product"]
		if ok {
			delete(builder, "spot_price_auto_product")
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAmazonSpotPriceProductDeprecation) Synopsis() string {
	return `Removes the deprecated "spot_price_auto_product" setting from Amazon builder templates`
}
