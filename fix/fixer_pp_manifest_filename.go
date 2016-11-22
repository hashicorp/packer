package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerManifestFilename renames any Filename to Output
type FixerManifestFilename struct{}

func (FixerManifestFilename) Fix(input map[string]interface{}) (map[string]interface{}, error) {

	// Our template type we'll use for this fixer only
	type template struct {
		PostProcessors []interface{} `mapstructure:"post-processors"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each post-processor and get out all the complex configs
	pps := make([]map[string]interface{}, 0, len(tpl.PostProcessors))
	for _, rawPP := range tpl.PostProcessors {
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

	for _, pp := range pps {
		ppTypeRaw, ok := pp["type"]
		if !ok {
			continue
		}

		if ppType, ok := ppTypeRaw.(string); !ok {
			continue
		} else if ppType != "manifest" {
			continue
		}

		filenameRaw, ok := pp["filename"]
		if !ok {
			continue
		}

		if filename, ok := filenameRaw.(string); ok {
			delete(pp, "filename")
			pp["output"] = filename
		}

	}

	input["post-processors"] = tpl.PostProcessors
	return input, nil
}

func (FixerManifestFilename) Synopsis() string {
	return `Updates "manifest" post-processor so any "filename" field is renamed to "output".`
}
