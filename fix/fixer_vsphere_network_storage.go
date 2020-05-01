package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVSphereNetworkStorage changes vsphere-iso network and networkCard fields into a network adapter and
// changes the disk_size, disk_thin_provisioned, and disk_eagerly_scrub into a storage adapter
type FixerVSphereNetworkStorage struct{}

func (FixerVSphereNetworkStorage) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		var networkAdapters []interface{}
		nic := make(map[string]interface{})
		hasNetwork := false

		networkRaw, ok := builder["network"]
		if ok {
			networkRawString, ok := networkRaw.(string)
			if !ok {
				// TODO: error?
				continue
			}
			nic["network"] = networkRawString
			delete(builder, "network")
			hasNetwork = true
		}

		networkCardRaw, ok := builder["networkCard"]
		if ok {
			networkCardRawString, ok := networkCardRaw.(string)
			if !ok {
				// TODO: error?
				continue
			}
			nic["networkCard"] = networkCardRawString
			delete(builder, "networkCard")
			hasNetwork = true
		}

		if hasNetwork {
			networkAdapters = append(networkAdapters, nic)
			adaptersRaw, ok := builder["network_adapters"]
			if ok {
				existingAdapters := adaptersRaw.([]interface{})
				networkAdapters = append(networkAdapters, existingAdapters...)
			}

			builder["network_adapters"] = networkAdapters
		}

		var storage []interface{}
		disk := make(map[string]interface{})
		hasStorage := false

		diskSizeRaw, ok := builder["disk_size"]
		if ok {
			diskSizeRawInt64, ok := diskSizeRaw.(int64)
			if !ok {
				// TODO: error?
				continue
			}
			disk["disk_size"] = diskSizeRawInt64
			delete(builder, "disk_size")
			hasStorage = true
		}

		discThinProvisionedRaw, ok := builder["disk_thin_provisioned"]
		if ok {
			discThinProvisionedRawBool, ok := discThinProvisionedRaw.(bool)
			if !ok {
				// TODO: error?
				continue
			}
			disk["disk_thin_provisioned"] = discThinProvisionedRawBool
			hasStorage = true
			delete(builder, "disk_thin_provisioned")
		}

		diskEagerlyScrubRaw, ok := builder["disk_eagerly_scrub"]
		if ok {
			diskEagerlyScrubRawBool, ok := diskEagerlyScrubRaw.(bool)
			if !ok {
				// TODO: error?
				continue
			}
			disk["disk_eagerly_scrub"] = diskEagerlyScrubRawBool
			hasStorage = true
			delete(builder, "disk_eagerly_scrub")
		}

		if hasStorage {
			storage = append(storage, disk)
			storageRaw, ok := builder["storage"]
			if ok {
				existingStorage := storageRaw.([]interface{})
				storage = append(storage, existingStorage...)
			}

			builder["storage"] = storage
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVSphereNetworkStorage) Synopsis() string {
	return `Updates "vmware" builders to "vmware-iso"`
}
