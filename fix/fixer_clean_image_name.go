// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"fmt"
	"regexp"

	"github.com/mitchellh/mapstructure"
)

// FixerCleanImageName is a Fixer that replaces the "clean_(image|ami)_name" template
// calls with "clean_resource_name"
type FixerCleanImageName struct{}

func (FixerCleanImageName) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*amazon*":             []string{"clean_ami_name"},
		"packer.googlecompute": []string{"clean_image_name"},
		"Azure*":               []string{"clean_image_name"},
	}
}

func (FixerCleanImageName) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`clean_(image|ami)_name`)

	// Go through each builder and replace CreateTime if we can
	for _, builder := range tpl.Builders {
		for key, value := range builder {
			switch v := value.(type) {
			case string:
				changed := re.ReplaceAllString(v, "clean_resource_name")
				builder[key] = changed
			case map[string]string:
				for k := range v {
					v[k] = re.ReplaceAllString(v[k], "clean_resource_name")
				}
				builder[key] = v
			case map[string]interface{}:
				for k := range v {
					if s, ok := v[k].(string); ok {
						v[k] = re.ReplaceAllString(s, "clean_resource_name")
					}
				}
				builder[key] = v
			default:
				if key == "image_labels" {

					panic(fmt.Sprintf("value: %#v", value))
				}
			}
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerCleanImageName) Synopsis() string {
	return `Replaces /clean_(image|ami)_name/ in builder configs with "clean_resource_name"`
}
