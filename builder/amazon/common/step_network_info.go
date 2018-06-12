package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepNetworkInfo queries AWS for information about
// VPC's, Subnets, and Security Groups that is used
// throughout the AMI creation process.
//
// Produces:
//   vpc_id string - the VPC ID
//   subnet_id string - the Subnet ID
//   az string - the AZ name
//   sg_ids []string - the SG IDs
type StepNetworkInfo struct {
	VpcId            string
	VpcFilter        VpcFilterOptions
	SubnetId         string
	SubnetFilter     SubnetFilterOptions
	AvailabilityZone string
	// TODO Security groups + filter
}

type subnetsSort []*ec2.Subnet

func (a subnetsSort) Len() int      { return len(a) }
func (a subnetsSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a subnetsSort) Less(i, j int) bool {
	return *a[i].AvailableIpAddressCount < *a[j].AvailableIpAddressCount
}

// Returns the most recent AMI out of a slice of images.
func mostFreeSubnet(subnets []*ec2.Subnet) *ec2.Subnet {
	sortedSubnets := subnets
	sort.Sort(subnetsSort(sortedSubnets))
	return sortedSubnets[len(sortedSubnets)-1]
}

func (s *StepNetworkInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// VPC
	if s.VpcId == "" && !s.VpcFilter.Empty() {
		params := &ec2.DescribeVpcsInput{}

		params.Filters = buildEc2Filters(s.VpcFilter.Filters)
		log.Printf("Using VPC Filters %v", params)

		vpcResp, err := ec2conn.DescribeVpcs(params)
		if err != nil {
			err := fmt.Errorf("Error querying VPCs: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(vpcResp.Vpcs) != 1 {
			err := fmt.Errorf("No or more than one VPC was found matching filters: %v", params)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.VpcId = *vpcResp.Vpcs[0].VpcId
		ui.Message(fmt.Sprintf("Found VPC ID: %s", s.VpcId))
	}

	// Subnet
	if s.SubnetId == "" && !s.SubnetFilter.Empty() {
		params := &ec2.DescribeSubnetsInput{}

		vpcId := "vpc-id"
		s.SubnetFilter.Filters[&vpcId] = &s.VpcId
		if s.AvailabilityZone != "" {
			az := "availability-zone"
			s.SubnetFilter.Filters[&az] = &s.AvailabilityZone
		}
		params.Filters = buildEc2Filters(s.SubnetFilter.Filters)
		log.Printf("Using Subnet Filters %v", params)

		subnetsResp, err := ec2conn.DescribeSubnets(params)
		if err != nil {
			err := fmt.Errorf("Error querying Subnets: %s", err)
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
			err := fmt.Errorf("Your query returned more than one result. Please try a more specific search, or set random or most_free to true.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		var subnet *ec2.Subnet
		switch {
		case s.SubnetFilter.MostFree:
			subnet = mostFreeSubnet(subnetsResp.Subnets)
		case s.SubnetFilter.Random:
			subnet = subnetsResp.Subnets[rand.Intn(len(subnetsResp.Subnets))]
		default:
			subnet = subnetsResp.Subnets[0]
		}
		s.SubnetId = *subnet.SubnetId
		ui.Message(fmt.Sprintf("Found Subnet ID: %s", s.SubnetId))
	}

	state.Put("vpc_id", s.VpcId)
	state.Put("subnet_id", s.SubnetId)
	return multistep.ActionContinue
}

func (s *StepNetworkInfo) Cleanup(multistep.StateBag) {}
