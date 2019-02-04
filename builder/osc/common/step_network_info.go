package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepNetworkInfo queries OUTSCALE for information about
// NET's and Subnets that is used throughout the OMI creation process.
//
// Produces (adding them to the state bag):
//   vpc_id string - the NET ID
//   subnet_id string - the Subnet ID
//   availability_zone string - the Subregion name
type StepNetworkInfo struct {
	NetId               string
	NetFilter           NetFilterOptions
	SubnetId            string
	SubnetFilter        SubnetFilterOptions
	SubregionName       string
	SecurityGroupIds    []string
	SecurityGroupFilter SecurityGroupFilterOptions
}

type subnetsSort []oapi.Subnet

func (a subnetsSort) Len() int      { return len(a) }
func (a subnetsSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a subnetsSort) Less(i, j int) bool {
	return a[i].AvailableIpsCount < a[j].AvailableIpsCount
}

// Returns the most recent OMI out of a slice of images.
func mostFreeSubnet(subnets []oapi.Subnet) oapi.Subnet {
	sortedSubnets := subnets
	sort.Sort(subnetsSort(sortedSubnets))
	return sortedSubnets[len(sortedSubnets)-1]
}

func (s *StepNetworkInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	// NET
	if s.NetId == "" && !s.NetFilter.Empty() {
		params := oapi.ReadNetsRequest{}
		params.Filters = buildNetFilters(s.NetFilter.Filters)
		s.NetFilter.Filters["state"] = "available"

		log.Printf("Using NET Filters %v", params)

		vpcResp, err := oapiconn.POST_ReadNets(params)
		if err != nil {
			err := fmt.Errorf("Error querying NETs: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(vpcResp.OK.Nets) != 1 {
			err := fmt.Errorf("Exactly one NET should match the filter, but %d NET's was found matching filters: %v", len(vpcResp.OK.Nets), params)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.NetId = vpcResp.OK.Nets[0].NetId
		ui.Message(fmt.Sprintf("Found NET ID: %s", s.NetId))
	}

	// Subnet
	if s.SubnetId == "" && !s.SubnetFilter.Empty() {
		params := oapi.ReadSubnetsRequest{}
		s.SubnetFilter.Filters["state"] = "available"

		if s.NetId != "" {
			s.SubnetFilter.Filters["vpc-id"] = s.NetId
		}
		if s.SubregionName != "" {
			s.SubnetFilter.Filters["availability-zone"] = s.SubregionName
		}
		params.Filters = buildSubnetFilters(s.SubnetFilter.Filters)
		log.Printf("Using Subnet Filters %v", params)

		subnetsResp, err := oapiconn.POST_ReadSubnets(params)
		if err != nil {
			err := fmt.Errorf("Error querying Subnets: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(subnetsResp.OK.Subnets) == 0 {
			err := fmt.Errorf("No Subnets was found matching filters: %v", params)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(subnetsResp.OK.Subnets) > 1 && !s.SubnetFilter.Random && !s.SubnetFilter.MostFree {
			err := fmt.Errorf("Your filter matched %d Subnets. Please try a more specific search, or set random or most_free to true.", len(subnetsResp.OK.Subnets))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		var subnet oapi.Subnet
		switch {
		case s.SubnetFilter.MostFree:
			subnet = mostFreeSubnet(subnetsResp.OK.Subnets)
		case s.SubnetFilter.Random:
			subnet = subnetsResp.OK.Subnets[rand.Intn(len(subnetsResp.OK.Subnets))]
		default:
			subnet = subnetsResp.OK.Subnets[0]
		}
		s.SubnetId = subnet.SubnetId
		ui.Message(fmt.Sprintf("Found Subnet ID: %s", s.SubnetId))
	}

	// Try to find Subregion and NET Id from Subnet if they are not yet found/given
	if s.SubnetId != "" && (s.SubregionName == "" || s.NetId == "") {
		log.Printf("[INFO] Finding Subregion and NetId for the given subnet '%s'", s.SubnetId)
		resp, err := oapiconn.POST_ReadSubnets(
			oapi.ReadSubnetsRequest{
				Filters: oapi.FiltersSubnet{
					SubnetIds: []string{s.SubnetId},
				},
			})
		if err != nil {
			err := fmt.Errorf("Describing the subnet: %s returned error: %s.", s.SubnetId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if s.SubregionName == "" {
			s.SubregionName = resp.OK.Subnets[0].SubregionName
			log.Printf("[INFO] SubregionName found: '%s'", s.SubregionName)
		}
		if s.NetId == "" {
			s.NetId = resp.OK.Subnets[0].NetId
			log.Printf("[INFO] NetId found: '%s'", s.NetId)
		}
	}

	state.Put("net_id", s.NetId)
	state.Put("subregion_name", s.SubregionName)
	state.Put("subnet_id", s.SubnetId)
	return multistep.ActionContinue
}

func (s *StepNetworkInfo) Cleanup(multistep.StateBag) {}
