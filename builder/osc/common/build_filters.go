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

func buildSubnetFilters(input map[string]string) oapi.FiltersSubnet {
	var filters oapi.FiltersSubnet
	for k, v := range input {
		filterValue := []string{v}
		switch name := k; name {
		case "available-ips-counts":
			if ipCount, err := strconv.Atoi(v); err == nil {
				filters.AvailableIpsCounts = []int64{int64(ipCount)}
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

func buildOMIFilters(input map[string]string) oapi.FiltersImage {
	var filters oapi.FiltersImage
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
		case "image-type":
			filters.ImageTypes = filterValue
		case "virtualization-type":
			filters.VirtualizationTypes = filterValue
		case "root-device-type":
			filters.RootDeviceTypes = filterValue
		case "block-device-mapping-volume-type":
			filters.BlockDeviceMappingVolumeType = filterValue
		//Some params are missing.
		default:
			log.Printf("[WARN] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}

func buildSecurityGroupFilters(input map[string]string) oapi.FiltersSecurityGroup {
	var filters oapi.FiltersSecurityGroup
	for k, v := range input {
		filterValue := []string{v}

		switch name := k; name {
		case "account-ids":
			filters.AccountIds = filterValue
		case "descriptions":
			filters.Descriptions = filterValue
		case "inbound-rule-account-ids":
			filters.InboundRuleAccountIds = filterValue
		case "inbound-rule-from-port-ranges":
			if val, err := strconv.Atoi(v); err == nil {
				filters.InboundRuleFromPortRanges = []int64{int64(val)}
			}
		case "inbound-rule-ip-ranges":
			filters.InboundRuleIpRanges = filterValue
		case "inbound-rule-protocols":
			filters.InboundRuleProtocols = filterValue
		case "inbound-rule-security-group-ids":
			filters.InboundRuleSecurityGroupIds = filterValue
		case "inbound-rule-security-group-names":
			filters.InboundRuleSecurityGroupNames = filterValue
		case "inbound-rule-to-port-ranges":
			if val, err := strconv.Atoi(v); err == nil {
				filters.InboundRuleToPortRanges = []int64{int64(val)}
			}
		case "net-ids":
			filters.NetIds = filterValue

		case "outbound-rule-account-ids":
			filters.OutboundRuleAccountIds = filterValue
		case "outbound-rule-from-port-ranges":
			if val, err := strconv.Atoi(v); err == nil {
				filters.OutboundRuleFromPortRanges = []int64{int64(val)}
			}
		case "outbound-rule-ip-ranges":
			filters.OutboundRuleIpRanges = filterValue
		case "outbound-rule-protocols":
			filters.OutboundRuleProtocols = filterValue
		case "outbound-rule-security-group-ids":
			filters.OutboundRuleSecurityGroupIds = filterValue
		case "outbound-rule-security-group-names":
			filters.OutboundRuleSecurityGroupNames = filterValue
		case "outbound-rule-to-port-ranges":
			if val, err := strconv.Atoi(v); err == nil {
				filters.OutboundRuleToPortRanges = []int64{int64(val)}
			}
		case "security-group-ids":
			filters.SecurityGroupIds = filterValue
		case "security-group-names":
			filters.SecurityGroupNames = filterValue
		case "tags-keys":
			filters.TagKeys = filterValue
		case "tags-values":
			filters.TagValues = filterValue
		//Some params are missing.
		default:
			log.Printf("[Debug] Unknown Filter Name: %s.", name)
		}
	}
	return filters
}
