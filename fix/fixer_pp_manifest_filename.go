// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerManifestFilename renames any Filename to Output
type FixerManifestFilename struct{}

func (FixerManifestFilename) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"packer.post-processor.manifest": []string{"filename"},
	}
}

func (FixerManifestFilename) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	if input["post-processors"] == nil {
		return input, nil
	}

	// Our template type we'll use for this fixer only
	type template struct {
		PP `mapstructure:",squash"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each post-processor and get out all the complex configs
	pps := tpl.ppList()

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
