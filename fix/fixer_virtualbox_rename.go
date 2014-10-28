package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVirtualBoxRename changes "virtualbox" builders to "virtualbox-iso"
type FixerVirtualBoxRename struct{}

func (FixerVirtualBoxRename) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "virtualbox" {
			continue
		}

		builder["type"] = "virtualbox-iso"
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVirtualBoxRename) Synopsis() string {
	return `Updates "virtualbox" builders to "virtualbox-iso"`
}
