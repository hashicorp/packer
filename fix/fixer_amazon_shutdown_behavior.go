// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerAmazonShutdownBehavior fix the spelling of "shutdown_behavior"
// template in a Amazon builder
type FixerAmazonShutdownBehavior struct{}

func (FixerAmazonShutdownBehavior) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*amazon*": []string{"shutdown_behaviour"},
	}
}

func (FixerAmazonShutdownBehavior) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if !strings.HasPrefix(builderType, "amazon-") {
			continue
		}

		shutdownBehavior, ok := builder["shutdown_behaviour"]

		if ok {
			builder["shutdown_behavior"] = shutdownBehavior
			delete(builder, "shutdown_behaviour")
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAmazonShutdownBehavior) Synopsis() string {
	return `Changes "shutdown_behaviour" to "shutdown_behavior" in Amazon builders.`
}
