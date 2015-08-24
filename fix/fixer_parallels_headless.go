package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerParallelsHeadless removes "headless" from a template in a Parallels builder
type FixerParallelsHeadless struct{}

func (FixerParallelsHeadless) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "parallels-iso" && builderType != "parallels-pvm" {
			continue
		}

		_, ok = builder["headless"]
		if !ok {
			continue
		}

		delete(builder, "headless")
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerParallelsHeadless) Synopsis() string {
	return `Removes unused "headless" from Parallels builders`
}
