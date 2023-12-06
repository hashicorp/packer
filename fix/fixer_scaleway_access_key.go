// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerScalewayAccessKey changes the "access_key" of a template
// to "organization_id".
type FixerScalewayAccessKey struct{}

func (FixerScalewayAccessKey) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"hashicorp.scaleway": []string{"access_key"},
	}
}

func (FixerScalewayAccessKey) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
		if builder["type"] != "scaleway" {
			continue
		}

		keyRaw, ok := builder["access_key"]
		if !ok {
			continue
		}

		accessKey, ok := keyRaw.(string)
		if !ok {
			continue
		}

		// only assign to organization_id if it doesn't
		// already exist; otherwise we'll just ignore access_key
		_, organizationIdIncluded := builder["organization_id"]
		if !organizationIdIncluded {
			builder["organization_id"] = accessKey
		}

		delete(builder, "access_key")
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerScalewayAccessKey) Synopsis() string {
	return `Updates builders using "access_key" to use "organization_id"`
}
