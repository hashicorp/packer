package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	retry "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepRunSpotInstance struct {
	AssociatePublicIpAddress          bool
	AvailabilityZone                  string
	BlockDevices                      BlockDevices
	Debug                             bool
	EbsOptimized                      bool
	ExpectedRootDevice                string
	IamInstanceProfile                string
	InstanceInitiatedShutdownBehavior string
	InstanceType                      string
	SourceAMI                         string
	SpotPrice                         string
	SpotPriceProduct                  string
	SubnetId                          string
	Tags                              TagMap
	VolumeTags                        TagMap
	UserData                          string
	UserDataFile                      string
	Ctx                               interpolate.Context

	instanceId  string
	spotRequest *ec2.SpotInstanceRequest
}

func (s *StepRunSpotInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	var keyName string
	if name, ok := state.GetOk("keyPair"); ok {
		keyName = name.(string)
	}
	securityGroupIds := aws.StringSlice(state.Get("securityGroupIds").([]string))
	ui := state.Get("ui").(packer.Ui)

	userData := s.UserData
	if s.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem reading user data file: %s", err))
			return multistep.ActionHalt
		}

		userData = string(contents)
	}

	// Test if it is encoded already, and if not, encode it
	if _, err := base64.StdEncoding.DecodeString(userData); err != nil {
		log.Printf("[DEBUG] base64 encoding user data...")
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}

	ui.Say("Launching a source AWS instance...")
	image, ok := state.Get("source_image").(*ec2.Image)
	if !ok {
		state.Put("error", fmt.Errorf("source_image type assertion failed"))
		return multistep.ActionHalt
	}
	s.SourceAMI = *image.ImageId

	if s.ExpectedRootDevice != "" && *image.RootDeviceType != s.ExpectedRootDevice {
		state.Put("error", fmt.Errorf(
			"The provided source AMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, *image.RootDeviceType))
		return multistep.ActionHalt
	}

	spotPrice := s.SpotPrice
	availabilityZone := s.AvailabilityZone
	if spotPrice == "auto" {
		ui.Message(fmt.Sprintf(
			"Finding spot price for %s %s...",
			s.SpotPriceProduct, s.InstanceType))

		// Detect the spot price
		startTime := time.Now().Add(-1 * time.Hour)
		resp, err := ec2conn.DescribeSpotPriceHistory(&ec2.DescribeSpotPriceHistoryInput{
			InstanceTypes:       []*string{&s.InstanceType},
			ProductDescriptions: []*string{&s.SpotPriceProduct},
			AvailabilityZone:    &s.AvailabilityZone,
			StartTime:           &startTime,
		})
		if err != nil {
			err := fmt.Errorf("Error finding spot price: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		var price float64
		for _, history := range resp.SpotPriceHistory {
			log.Printf("[INFO] Candidate spot price: %s", *history.SpotPrice)
			current, err := strconv.ParseFloat(*history.SpotPrice, 64)
			if err != nil {
				log.Printf("[ERR] Error parsing spot price: %s", err)
				continue
			}
			if price == 0 || current < price {
				price = current
				if s.AvailabilityZone == "" {
					availabilityZone = *history.AvailabilityZone
				}
			}
		}
		if price == 0 {
			err := fmt.Errorf("No candidate spot prices found!")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			// Add 0.5 cents to minimum spot bid to ensure capacity will be available
			// Avoids price-too-low error in active markets which can fluctuate
			price = price + 0.005
		}

		spotPrice = strconv.FormatFloat(price, 'f', -1, 64)
	}

	var instanceId string

	ui.Say("Adding tags to source instance")
	if _, exists := s.Tags["Name"]; !exists {
		s.Tags["Name"] = "Packer Builder"
	}

	ec2Tags, err := s.Tags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ec2Tags.Report(ui)

	ui.Message(fmt.Sprintf(
		"Requesting spot instance '%s' for: %s",
		s.InstanceType, spotPrice))

	runOpts := &ec2.RequestSpotLaunchSpecification{
		ImageId:            &s.SourceAMI,
		InstanceType:       &s.InstanceType,
		UserData:           &userData,
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Name: &s.IamInstanceProfile},
		Placement: &ec2.SpotPlacement{
			AvailabilityZone: &availabilityZone,
		},
		BlockDeviceMappings: s.BlockDevices.BuildLaunchDevices(),
		EbsOptimized:        &s.EbsOptimized,
	}

	if s.SubnetId != "" && s.AssociatePublicIpAddress {
		runOpts.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: &s.AssociatePublicIpAddress,
				SubnetId:                 &s.SubnetId,
				Groups:                   securityGroupIds,
				DeleteOnTermination:      aws.Bool(true),
			},
		}
	} else {
		runOpts.SubnetId = &s.SubnetId
		runOpts.SecurityGroupIds = securityGroupIds
	}

	if keyName != "" {
		runOpts.KeyName = &keyName
	}

	runSpotResp, err := ec2conn.RequestSpotInstances(&ec2.RequestSpotInstancesInput{
		SpotPrice:           &spotPrice,
		LaunchSpecification: runOpts,
	})
	if err != nil {
		err := fmt.Errorf("Error launching source spot instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.spotRequest = runSpotResp.SpotInstanceRequests[0]

	spotRequestId := s.spotRequest.SpotInstanceRequestId
	ui.Message(fmt.Sprintf("Waiting for spot request (%s) to become active...", *spotRequestId))
	err = WaitUntilSpotRequestFulfilled(ctx, ec2conn, *spotRequestId)
	if err != nil {
		err := fmt.Errorf("Error waiting for spot request (%s) to become ready: %s", *spotRequestId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	spotResp, err := ec2conn.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{spotRequestId},
	})
	if err != nil {
		err := fmt.Errorf("Error finding spot request (%s): %s", *spotRequestId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instanceId = *spotResp.SpotInstanceRequests[0].InstanceId

	// Set the instance ID so that the cleanup works properly
	s.instanceId = instanceId

	ui.Message(fmt.Sprintf("Instance ID: %s", instanceId))
	ui.Say(fmt.Sprintf("Waiting for instance (%v) to become ready...", instanceId))
	describeInstance := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	if err := ec2conn.WaitUntilInstanceRunningWithContext(ctx, describeInstance); err != nil {
		err := fmt.Errorf("Error waiting for instance (%s) to become ready: %s", instanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	r, err := ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	})
	if err != nil || len(r.Reservations) == 0 || len(r.Reservations[0].Instances) == 0 {
		err := fmt.Errorf("Error finding source instance.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instance := r.Reservations[0].Instances[0]

	// Retry creating tags for about 2.5 minutes
	err = retry.Retry(0.2, 30, 11, func(_ uint) (bool, error) {
		_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
			Tags:      ec2Tags,
			Resources: []*string{instance.InstanceId},
		})
		if err == nil {
			return true, nil
		}
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidInstanceID.NotFound" {
				return false, nil
			}
		}
		return true, err
	})

	if err != nil {
		err := fmt.Errorf("Error tagging source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	volumeIds := make([]*string, 0)
	for _, v := range instance.BlockDeviceMappings {
		if ebs := v.Ebs; ebs != nil {
			volumeIds = append(volumeIds, ebs.VolumeId)
		}
	}

	if len(volumeIds) > 0 && s.VolumeTags.IsSet() {
		ui.Say("Adding tags to source EBS Volumes")

		volumeTags, err := s.VolumeTags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
		if err != nil {
			err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		volumeTags.Report(ui)

		_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
			Resources: volumeIds,
			Tags:      volumeTags,
		})

		if err != nil {
			err := fmt.Errorf("Error tagging source EBS Volumes on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	if s.Debug {
		if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
			ui.Message(fmt.Sprintf("Public DNS: %s", *instance.PublicDnsName))
		}

		if instance.PublicIpAddress != nil && *instance.PublicIpAddress != "" {
			ui.Message(fmt.Sprintf("Public IP: %s", *instance.PublicIpAddress))
		}

		if instance.PrivateIpAddress != nil && *instance.PrivateIpAddress != "" {
			ui.Message(fmt.Sprintf("Private IP: %s", *instance.PrivateIpAddress))
		}
	}

	state.Put("instance", instance)

	return multistep.ActionContinue
}

func (s *StepRunSpotInstance) Cleanup(state multistep.StateBag) {

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// Cancel the spot request if it exists
	if s.spotRequest != nil {
		ui.Say("Cancelling the spot request...")
		input := &ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{s.spotRequest.SpotInstanceRequestId},
		}
		if _, err := ec2conn.CancelSpotInstanceRequests(input); err != nil {
			ui.Error(fmt.Sprintf("Error cancelling the spot request, may still be around: %s", err))
			return
		}

		err := WaitUntilSpotRequestFulfilled(aws.BackgroundContext(), ec2conn, *s.spotRequest.SpotInstanceRequestId)
		if err != nil {
			ui.Error(err.Error())
		}

	}

	// Terminate the source instance if it exists
	if s.instanceId != "" {
		ui.Say("Terminating the source AWS instance...")
		if _, err := ec2conn.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: []*string{&s.instanceId}}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
			return
		}

		if err := WaitUntilInstanceTerminated(aws.BackgroundContext(), ec2conn, s.instanceId); err != nil {
			ui.Error(err.Error())
		}
	}
}
