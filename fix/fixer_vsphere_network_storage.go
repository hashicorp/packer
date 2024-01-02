// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVSphereNetworkDisk changes vsphere-iso network and networkCard fields
// into a network adapter and changes the disk_size, disk_thin_provisioned, and
// disk_eagerly_scrub into a storage adapter
type FixerVSphereNetworkDisk struct{}

func (FixerVSphereNetworkDisk) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"jetbrains.vsphere": []string{"network_card", "network", "networkCard", "disk_size",
			"disk_thin_provisioned", "disk_eagerly_scrub"},
	}
}

func (FixerVSphereNetworkDisk) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
			nic["network"] = networkRaw
			delete(builder, "network")
			hasNetwork = true
		}

		// legacy syntax from when VSphere was 3rd party
		networkCardRaw, ok := builder["networkCard"]
		if ok {
			nic["network_card"] = networkCardRaw
			delete(builder, "networkCard")
			hasNetwork = true
		}

		// underscored syntax used when Packer merged vSphere
		networkCardRaw, ok = builder["network_card"]
		if ok {
			nic["network_card"] = networkCardRaw
			delete(builder, "network_card")
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
			disk["disk_size"] = diskSizeRaw
			delete(builder, "disk_size")
			hasStorage = true
		}

		discThinProvisionedRaw, ok := builder["disk_thin_provisioned"]
		if ok {
			disk["disk_thin_provisioned"] = discThinProvisionedRaw
			hasStorage = true
			delete(builder, "disk_thin_provisioned")
		}

		diskEagerlyScrubRaw, ok := builder["disk_eagerly_scrub"]
		if ok {
			disk["disk_eagerly_scrub"] = diskEagerlyScrubRaw
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

func (FixerVSphereNetworkDisk) Synopsis() string {
	return `Removes deprecated network and disk fields from "vsphere-iso" builder`
}
