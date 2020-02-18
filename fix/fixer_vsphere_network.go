package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVSphereNetwork changes vsphere-iso network and networkCard fields into a network Adapter
type FixerVSphereNetwork struct{}

func (FixerVSphereNetwork) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "vsphere-iso" {
			continue
		}

		networkRaw, ok := builder["network"]
		if !ok {
			continue
		}
		networkRawString, ok := networkRaw.(string)
		if !ok {
			// TODO: error?
			continue
		}

		delete(builder, "network")
		networkCardRaw, ok := builder["networkCard"]
		if !ok {
			continue
		}
		networkCardRawString, ok := networkCardRaw.(string)
		if !ok {
			// TODO: error?
			continue
		}
		delete(builder, "networkCard")

		var networkAdapters []interface{}
		nic := make(map[string]interface{})
		nic["network"] = networkRawString
		nic["networkCard"] = networkCardRawString

		networkAdapters = append(networkAdapters, nic)
		adaptersRaw, ok := builder["network_adapters"]
		if ok {
			existingAdapters := adaptersRaw.([]interface{})
			networkAdapters = append(networkAdapters, existingAdapters...)
		}

		builder["network_adapters"] = networkAdapters
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVSphereNetwork) Synopsis() string {
	return `Updates "vmware" builders to "vmware-iso"`
}
