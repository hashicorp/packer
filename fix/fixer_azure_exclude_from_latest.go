// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerAzureExcludeFromLatest fix the spelling of "exclude_from_latest"
// template in an Azure builder
type FixerAzureExcludeFromLatest struct{}

func (FixerAzureExcludeFromLatest) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"Azure*": []string{"exlude_from_latest"},
	}
}

func (FixerAzureExcludeFromLatest) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if !strings.HasPrefix(builderType, "azure-chroot") {
			continue
		}

		if !strings.HasPrefix(builderType, "azure-chroot") {
			continue
		}

		sharedImageDestination, ok := builder["shared_image_destination"].(map[string]interface{})
		if !ok {
			continue
		}

		excludeFromLatest, ok := sharedImageDestination["exlude_from_latest"]
		if !ok {
			continue
		}

		sharedImageDestination["exclude_from_latest"] = excludeFromLatest
		delete(sharedImageDestination, "exlude_from_latest")

		builder["shared_image_destination"] = sharedImageDestination
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAzureExcludeFromLatest) Synopsis() string {
	return `Changes "exlude_from_latest" to "exclude_from_latest" in Azure builders.`
}
