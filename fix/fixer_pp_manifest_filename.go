package fix

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

// FixerManifestFilename renames any Filename to Output
type FixerManifestFilename struct{}

func (FixerManifestFilename) Fix(input map[string]interface{}) (map[string]interface{}, error) {

	// Our template type we'll use for this fixer only
	type template struct {
		PostProcessors []map[string]interface{} `mapstructure:"post-processors"`
	}

	// Decode the input into our structure, if we can
	fmt.Println("Got 0")
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		fmt.Println("Got 1")
		return nil, err
	}
	for _, pp := range tpl.PostProcessors {
		ppTypeRaw, ok := pp["type"]
		if !ok {
			continue
		}

		ppType, ok := ppTypeRaw.(string)
		if !ok {
			continue
		}

		if ppType != "manifest" {
			continue
		}

		filenameRaw, ok := pp["filename"]
		if !ok {
			continue
		}

		filename, ok := filenameRaw.(string)
		if !ok {
			continue
		}

		delete(pp, "filename")
		pp["output"] = filename
	}

	input["post-processors"] = tpl.PostProcessors
	return input, nil
}

func (FixerManifestFilename) Synopsis() string {
	return `Updates "manifest" post-processor so any "filename" field is renamed to "output".`
}
