// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

type FixerAmazonTemporarySecurityCIDRs struct{}

func (FixerAmazonTemporarySecurityCIDRs) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*amazon*": []string{"temporary_security_group_source_cidr"},
	}
}

func (FixerAmazonTemporarySecurityCIDRs) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each builder and replace the temporary_security_group_source_cidr if we can
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

		temporarySecurityGroupCIDR, ok := builder["temporary_security_group_source_cidr"].(string)
		if !ok {
			continue
		}

		delete(builder, "temporary_security_group_source_cidr")
		builder["temporary_security_group_source_cidrs"] = []string{temporarySecurityGroupCIDR}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAmazonTemporarySecurityCIDRs) Synopsis() string {
	return `Replaces "temporary_security_group_source_cidr" (string) with "temporary_security_group_source_cidrs" (list of strings)`
}
