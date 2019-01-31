package common

import (
	"log"

	"github.com/outscale/osc-go/oapi"
)

func buildOMIFilters(input map[string]string) oapi.FiltersImage {
	var filters oapi.FiltersImage
	for k, v := range input {
		filterValue := []string{v}

		switch name := k; name {
		case "account_aliases":
			filters.AccountAliases = filterValue
		case "account_ids":
			filters.AccountIds = filterValue
		case "architectures":
			filters.Architectures = filterValue
		case "image_ids":
			filters.ImageIds = filterValue
		case "image_names":
			filters.ImageNames = filterValue
		case "image_types":
			filters.ImageTypes = filterValue
		case "virtualization_types":
			filters.VirtualizationTypes = filterValue
		case "root_device_types":
			filters.RootDeviceTypes = filterValue
		case "block_device_mapping_volume_type":
			filters.BlockDeviceMappingVolumeType = filterValue
		//Some params are missing.
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}
