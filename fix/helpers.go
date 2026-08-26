// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package fix

// PP is a convenient way to interact with the post-processors within a fixer
type PP struct {
	PostProcessors []any `mapstructure:"post-processors"`
}

// postProcessors converts the variable structure of the template to a list
func (pp *PP) ppList() []map[string]any {
	pps := make([]map[string]any, 0, len(pp.PostProcessors))
	for _, rawPP := range pp.PostProcessors {
		switch pp := rawPP.(type) {
		case string:
		case map[string]any:
			pps = append(pps, pp)
		case []any:
			for _, innerRawPP := range pp {
				if innerPP, ok := innerRawPP.(map[string]any); ok {
					pps = append(pps, innerPP)
				}
			}
		}
	}
	return pps
}
