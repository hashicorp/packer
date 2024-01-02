// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"regexp"

	"github.com/mitchellh/mapstructure"
)

// FixerCreateTime is a Fixer that replaces the ".CreateTime" template
// calls with "{{timestamp}"
type FixerCreateTime struct{}

func (FixerCreateTime) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerCreateTime) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	badKeys := []string{
		"ami_name",
		"bundle_prefix",
		"snapshot_name",
	}

	re := regexp.MustCompile(`{{\s*\.CreateTime\s*}}`)

	// Go through each builder and replace CreateTime if we can
	for _, builder := range tpl.Builders {
		for _, key := range badKeys {
			raw, ok := builder[key]
			if !ok {
				continue
			}

			v, ok := raw.(string)
			if !ok {
				continue
			}

			builder[key] = re.ReplaceAllString(v, "{{timestamp}}")
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerCreateTime) Synopsis() string {
	return `Replaces ".CreateTime" in builder configs with "{{timestamp}}"`
}
