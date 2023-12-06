// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVMwareCompaction adds "skip_compaction = true" to "vmware-iso" builders with incompatible disk_type_id
type FixerVMwareCompaction struct{}

func (FixerVMwareCompaction) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerVMwareCompaction) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "vmware-iso" {
			continue
		}

		builderRemoteTypeRaw, ok := builder["remote_type"]
		if !ok {
			continue
		}

		builderRemoteType, ok := builderRemoteTypeRaw.(string)
		if !ok {
			continue
		}

		if builderRemoteType != "esx5" {
			continue
		}

		builderDiskTypeIdRaw, ok := builder["disk_type_id"]
		// set to default when this fixer was added due to incompatibility of defaults
		if !ok {
			builderDiskTypeId := "zeroedthick"
			builder["disk_type_id"] = builderDiskTypeId
		}

		if ok {
			builderDiskTypeId, ok := builderDiskTypeIdRaw.(string)
			if !ok {
				continue
			}
			if builderDiskTypeId == "thin" {
				continue
			}
		}

		builderSkipCompactionRaw, ok := builder["skip_compaction"]
		// already verified this is not creating a "thin" disk, will need to skip_compaction
		if ok {
			builderSkipCompaction, ok := builderSkipCompactionRaw.(bool)
			if !ok {
				continue
			}
			if !builderSkipCompaction {
				builder["skip_compaction"] = !builderSkipCompaction
			}
			continue
		}

		builderSkipCompaction := true
		builder["skip_compaction"] = builderSkipCompaction
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVMwareCompaction) Synopsis() string {
	return `Adds "skip_compaction = true" to "vmware-iso" builders with incompatible disk_type_id`
}
