package common

import (
	"log"
	"strconv"

	"github.com/outscale/osc-sdk-go/osc"
)

func buildOscNetFilters(input map[string]string) osc.FiltersNet {
	var filters osc.FiltersNet
	for k, v := range input {
		filterValue := []string{v}
		switch name := k; name {
		case "ip-range":
			filters.IpRanges = filterValue
		case "dhcp-options-set-id":
			filters.DhcpOptionsSetIds = filterValue
		case "is-default":
			if isDefault, err := strconv.ParseBool(v); err == nil {
				filters.IsDefault = isDefault
			}
		case "state":
			filters.States = filterValue
		case "tag-key":
			filters.TagKeys = filterValue
		case "tag-value":
			filters.TagValues = filterValue
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}

func buildOscSubnetFilters(input map[string]string) osc.FiltersSubnet {
	var filters osc.FiltersSubnet
	for k, v := range input {
		filterValue := []string{v}
		switch name := k; name {
		case "available-ips-counts":
			if ipCount, err := strconv.Atoi(v); err == nil {
				filters.AvailableIpsCounts = []int32{int32(ipCount)}
			}
		case "ip-ranges":
			filters.IpRanges = filterValue
		case "net-ids":
			filters.NetIds = filterValue
		case "states":
			filters.States = filterValue
		case "subnet-ids":
			filters.SubnetIds = filterValue
		case "sub-region-names":
			filters.SubregionNames = filterValue
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}

func buildOSCOMIFilters(input map[string]string) osc.FiltersImage {
	var filters osc.FiltersImage
	for k, v := range input {
		filterValue := []string{v}

		switch name := k; name {
		case "account-alias":
			filters.AccountAliases = filterValue
		case "account-id":
			filters.AccountIds = filterValue
		case "architecture":
			filters.Architectures = filterValue
		case "image-id":
			filters.ImageIds = filterValue
		case "image-name":
			filters.ImageNames = filterValue
		// case "image-type":
		// 	filters.ImageTypes = filterValue
		case "virtualization-type":
			filters.VirtualizationTypes = filterValue
		case "root-device-type":
			filters.RootDeviceTypes = filterValue
		// case "block-device-mapping-volume-type":
		// 	filters.BlockDeviceMappingVolumeType = filterValue
		//Some params are missing.
		default:
			log.Printf("[WARN] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}
