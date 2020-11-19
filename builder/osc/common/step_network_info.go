package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
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

type subnetsOscSort []osc.Subnet

func (a subnetsOscSort) Len() int      { return len(a) }
func (a subnetsOscSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a subnetsOscSort) Less(i, j int) bool {
	return a[i].AvailableIpsCount < a[j].AvailableIpsCount
}

// Returns the most recent OMI out of a slice of images.
func mostFreeOscSubnet(subnets []osc.Subnet) osc.Subnet {
	sortedSubnets := subnets
	sort.Sort(subnetsOscSort(sortedSubnets))
	return sortedSubnets[len(sortedSubnets)-1]
}

//Run ...
func (s *StepNetworkInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	ui := state.Get("ui").(packersdk.Ui)

	// NET
	if s.NetId == "" && !s.NetFilter.Empty() {
		params := osc.ReadNetsRequest{}
		params.Filters = buildOscNetFilters(s.NetFilter.Filters)
		s.NetFilter.Filters["state"] = "available"

		log.Printf("Using NET Filters %v", params)

		vpcResp, _, err := oscconn.NetApi.ReadNets(context.Background(), &osc.ReadNetsOpts{
			ReadNetsRequest: optional.NewInterface(params),
		})
		if err != nil {
			err := fmt.Errorf("Error querying NETs: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(vpcResp.Nets) != 1 {
			err := fmt.Errorf("Exactly one NET should match the filter, but %d NET's was found matching filters: %v", len(vpcResp.Nets), params)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.NetId = vpcResp.Nets[0].NetId
		ui.Message(fmt.Sprintf("Found NET ID: %s", s.NetId))
	}

	// Subnet
	if s.SubnetId == "" && !s.SubnetFilter.Empty() {
		params := osc.ReadSubnetsRequest{}
		s.SubnetFilter.Filters["state"] = "available"

		if s.NetId != "" {
			s.SubnetFilter.Filters["vpc-id"] = s.NetId
		}
		if s.SubregionName != "" {
			s.SubnetFilter.Filters["availability-zone"] = s.SubregionName
		}
		params.Filters = buildOscSubnetFilters(s.SubnetFilter.Filters)
		log.Printf("Using Subnet Filters %v", params)

		subnetsResp, _, err := oscconn.SubnetApi.ReadSubnets(context.Background(), &osc.ReadSubnetsOpts{
			ReadSubnetsRequest: optional.NewInterface(params),
		})
		if err != nil {
			err := fmt.Errorf("error querying Subnets: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(subnetsResp.Subnets) == 0 {
			err := fmt.Errorf("No Subnets was found matching filters: %v", params)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(subnetsResp.Subnets) > 1 && !s.SubnetFilter.Random && !s.SubnetFilter.MostFree {
			err := fmt.Errorf("your filter matched %d Subnets. Please try a more specific search, or set random or most_free to true", len(subnetsResp.Subnets))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		var subnet osc.Subnet
		switch {
		case s.SubnetFilter.MostFree:
			subnet = mostFreeOscSubnet(subnetsResp.Subnets)
		case s.SubnetFilter.Random:
			subnet = subnetsResp.Subnets[rand.Intn(len(subnetsResp.Subnets))]
		default:
			subnet = subnetsResp.Subnets[0]
		}
		s.SubnetId = subnet.SubnetId
		ui.Message(fmt.Sprintf("Found Subnet ID: %s", s.SubnetId))
	}

	// Try to find Subregion and NET Id from Subnet if they are not yet found/given
	if s.SubnetId != "" && (s.SubregionName == "" || s.NetId == "") {
		log.Printf("[INFO] Finding Subregion and NetId for the given subnet '%s'", s.SubnetId)
		resp, _, err := oscconn.SubnetApi.ReadSubnets(context.Background(), &osc.ReadSubnetsOpts{
			ReadSubnetsRequest: optional.NewInterface(osc.ReadSubnetsRequest{
				Filters: osc.FiltersSubnet{
					SubnetIds: []string{s.SubnetId},
				},
			}),
		})
		if err != nil {
			err := fmt.Errorf("describing the subnet: %s returned error: %s", s.SubnetId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if s.SubregionName == "" {
			s.SubregionName = resp.Subnets[0].SubregionName
			log.Printf("[INFO] SubregionName found: '%s'", s.SubregionName)
		}
		if s.NetId == "" {
			s.NetId = resp.Subnets[0].NetId
			log.Printf("[INFO] NetId found: '%s'", s.NetId)
		}
	}

	state.Put("net_id", s.NetId)
	state.Put("subregion_name", s.SubregionName)
	state.Put("subnet_id", s.SubnetId)
	return multistep.ActionContinue
}

//Cleanup ...
func (s *StepNetworkInfo) Cleanup(multistep.StateBag) {}
