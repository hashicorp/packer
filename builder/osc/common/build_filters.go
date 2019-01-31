package common

import (
	"log"
	"strconv"

	"github.com/outscale/osc-go/oapi"
)

func buildNetFilters(input map[string]string) oapi.FiltersNet {
	var filters oapi.FiltersNet
	for k, v := range input {
		filterValue := []string{v}
		switch name := k; name {
		case "ip_range":
			filters.IpRanges = filterValue
		case "dhcp_options_set_id":
			filters.DhcpOptionsSetIds = filterValue
		case "is_default":
			if isDefault, err := strconv.ParseBool(v); err == nil {
				filters.IsDefault = isDefault
			}
		case "state":
			filters.States = filterValue
		case "tag_key":
			filters.TagKeys = filterValue
		case "tag_value":
			filters.TagValues = filterValue
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}

func buildSubnetFilters(input map[string]string) oapi.FiltersSubnet {
	var filters oapi.FiltersSubnet
	for k, v := range input {
		filterValue := []string{v}
		switch name := k; name {
		case "available_ips_counts":
			if ipCount, err := strconv.Atoi(v); err == nil {
				filters.AvailableIpsCounts = []int64{int64(ipCount)}
			}
		case "ip_ranges":
			filters.IpRanges = filterValue
		case "net_ids":
			filters.NetIds = filterValue
		case "states":
			filters.States = filterValue
		case "subnet_ids":
			filters.SubnetIds = filterValue
		case "sub_region_names":
			filters.SubregionNames = filterValue
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}

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
