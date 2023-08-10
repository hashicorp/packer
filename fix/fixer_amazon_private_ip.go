// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerAmazonPrivateIP is a Fixer that replaces instances of `"private_ip":
// true` with `"ssh_interface": "private_ip"`
type FixerAmazonPrivateIP struct{}

func (FixerAmazonPrivateIP) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*amazon*": []string{"ssh_private_ip"},
	}
}

func (FixerAmazonPrivateIP) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		// if ssh_interface already set, do nothing
		if _, ok := builder["ssh_interface"]; ok {
			continue
		}

		privateIPi, ok := builder["ssh_private_ip"]
		if !ok {
			continue
		}
		privateIP, ok := privateIPi.(bool)
		if !ok {
			var err error
			privateIP, err = strconv.ParseBool(privateIPi.(string))
			if err != nil {
				return nil, fmt.Errorf("ssh_private_ip is not a boolean, %s", err)
			}
		}

		delete(builder, "ssh_private_ip")
		if privateIP {
			builder["ssh_interface"] = "private_ip"
		} else {
			builder["ssh_interface"] = "public_ip"
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerAmazonPrivateIP) Synopsis() string {
	return "Replaces `\"ssh_private_ip\": true` in amazon builders with `\"ssh_interface\": \"private_ip\"`"
}
