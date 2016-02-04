package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerParallelsDeprecations removes "parallels_tools_host_path" from a
// template in a Parallels builder and changes "guest_os_distribution" to
// "guest_os_type", possibly overwriting any existing "guest_os_type"
type FixerParallelsDeprecations struct{}

func (FixerParallelsDeprecations) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		_, ok = builder["parallels_tools_host_path"]
		if ok {
			delete(builder, "parallels_tools_host_path")
		}

		guestOsDistribution, ok := builder["guest_os_distribution"]

		if ok {
			builder["guest_os_type"] = guestOsDistribution
			delete(builder, "guest_os_distribution")
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerParallelsDeprecations) Synopsis() string {
	return `Removes deprecated "parallels_tools_host_path" from Parallels builders
	and changes "guest_os_distribution" to "guest_os_type".`
}
