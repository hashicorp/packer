package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepNetworkInfo queries AWS for information about
// VPC's and Subnets that is used throughout the AMI creation process.
//
// Produces (adding them to the state bag):
//   vpc_id string - the VPC ID
//   subnet_id string - the Subnet ID
//   availability_zone string - the AZ name
type StepNetworkInfo struct {
	VpcId               string
	VpcFilter           VpcFilterOptions
	SubnetId            string
	SubnetFilter        SubnetFilterOptions
	AvailabilityZone    string
	SecurityGroupIds    []string
	SecurityGroupFilter SecurityGroupFilterOptions
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

func (s *StepNetworkInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packersdk.Ui)

	// VPC
	if s.VpcId == "" && !s.VpcFilter.Empty() {
		params := &ec2.DescribeVpcsInput{}
		params.Filters = buildEc2Filters(s.VpcFilter.Filters)
		s.VpcFilter.Filters["state"] = "available"

		log.Printf("Using VPC Filters %v", params)

		vpcResp, err := ec2conn.DescribeVpcs(params)
		if err != nil {
			err := fmt.Errorf("Error querying VPCs: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(vpcResp.Vpcs) != 1 {
			err := fmt.Errorf("Exactly one VPC should match the filter, but %d VPC's was found matching filters: %v", len(vpcResp.Vpcs), params)
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
		s.SubnetFilter.Filters["state"] = "available"

		if s.VpcId != "" {
			s.SubnetFilter.Filters["vpc-id"] = s.VpcId
		}
		if s.AvailabilityZone != "" {
			s.SubnetFilter.Filters["availabilityZone"] = s.AvailabilityZone
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
			err := fmt.Errorf("Your filter matched %d Subnets. Please try a more specific search, or set random or most_free to true.", len(subnetsResp.Subnets))
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

	// Try to find AZ and VPC Id from Subnet if they are not yet found/given
	if s.SubnetId != "" && (s.AvailabilityZone == "" || s.VpcId == "") {
		log.Printf("[INFO] Finding AZ and VpcId for the given subnet '%s'", s.SubnetId)
		resp, err := ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: []*string{&s.SubnetId}})
		if err != nil {
			err := fmt.Errorf("Describing the subnet: %s returned error: %s.", s.SubnetId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if s.AvailabilityZone == "" {
			s.AvailabilityZone = *resp.Subnets[0].AvailabilityZone
			log.Printf("[INFO] AvailabilityZone found: '%s'", s.AvailabilityZone)
		}
		if s.VpcId == "" {
			s.VpcId = *resp.Subnets[0].VpcId
			log.Printf("[INFO] VpcId found: '%s'", s.VpcId)
		}
	}

	state.Put("vpc_id", s.VpcId)
	state.Put("availability_zone", s.AvailabilityZone)
	state.Put("subnet_id", s.SubnetId)
	return multistep.ActionContinue
}

func (s *StepNetworkInfo) Cleanup(multistep.StateBag) {}
