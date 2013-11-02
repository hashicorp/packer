package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVirtualBoxGAAttach changes the "guest_additions_attach" of a template
// to "guest_additions_mode".
type FixerVirtualBoxGAAttach struct{}

func (FixerVirtualBoxGAAttach) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		gaAttachRaw, ok := builder["guest_additions_attach"]
		if !ok {
			continue
		}

		gaAttach, ok := gaAttachRaw.(bool)
		if !ok {
			continue
		}

		gaMode := "upload"
		if gaAttach {
			gaMode = "attach"
		}

		delete(builder, "guest_additions_attach")
		builder["guest_additions_mode"] = gaMode
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVirtualBoxGAAttach) Synopsis() string {
	return `Updates VirtualBox builders using "guest_additions_attach" to use "guest_additions_mode"`
}
