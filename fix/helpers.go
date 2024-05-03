// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

// PP is a convenient way to interact with the post-processors within a fixer
type PP struct {
	PostProcessors []interface{} `mapstructure:"post-processors"`
}

// postProcessors converts the variable structure of the template to a list
func (pp *PP) ppList() []map[string]interface{} {
	pps := make([]map[string]interface{}, 0, len(pp.PostProcessors))
	for _, rawPP := range pp.PostProcessors {
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
	return pps
}
