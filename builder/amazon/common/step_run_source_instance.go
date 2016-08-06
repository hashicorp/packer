package common

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepRunSourceInstance struct {
	AssociatePublicIpAddress          bool
	AvailabilityZone                  string
	BlockDevices                      BlockDevices
	Debug                             bool
	EbsOptimized                      bool
	ExpectedRootDevice                string
	InstanceType                      string
	IamInstanceProfile                string
	SourceAMI                         string
	SpotPrice                         string
	SpotPriceProduct                  string
	SubnetId                          string
	Tags                              map[string]string
	UserData                          string
	UserDataFile                      string
	InstanceInitiatedShutdownBehavior string

	instanceId  string
	spotRequest *ec2.SpotInstanceRequest
}

func (s *StepRunSourceInstance) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	keyName := state.Get("keyPair").(string)
	tempSecurityGroupIds := state.Get("securityGroupIds").([]string)
	ui := state.Get("ui").(packer.Ui)

	securityGroupIds := make([]*string, len(tempSecurityGroupIds))
	for i, sg := range tempSecurityGroupIds {
		found := false
		for i := 0; i < 5; i++ {
			time.Sleep(time.Duration(i) * 5 * time.Second)
			log.Printf("[DEBUG] Describing tempSecurityGroup to ensure it is available: %s", sg)
			_, err := ec2conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
				GroupIds: []*string{aws.String(sg)},
			})
			if err == nil {
				log.Printf("[DEBUG] Found security group %s", sg)
				found = true
				break
			}
			log.Printf("[DEBUG] Error in querying security group %s", err)
		}
		if found {
			securityGroupIds[i] = aws.String(sg)
		} else {
			state.Put("error", fmt.Errorf("Timeout waiting for security group %s to become available", sg))
		}
	}

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
	imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{&s.SourceAMI},
	})
	if err != nil {
		state.Put("error", fmt.Errorf("There was a problem with the source AMI: %s", err))
		return multistep.ActionHalt
	}

	if len(imageResp.Images) != 1 {
		state.Put("error", fmt.Errorf("The source AMI '%s' could not be found.", s.SourceAMI))
		return multistep.ActionHalt
	}

	if s.ExpectedRootDevice != "" && *imageResp.Images[0].RootDeviceType != s.ExpectedRootDevice {
		state.Put("error", fmt.Errorf(
			"The provided source AMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, *imageResp.Images[0].RootDeviceType))
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
		}

		spotPrice = strconv.FormatFloat(price, 'f', -1, 64)
	}

	var instanceId string

	if spotPrice == "" || spotPrice == "0" {
		runOpts := &ec2.RunInstancesInput{
			KeyName:                           &keyName,
			ImageId:                           &s.SourceAMI,
			InstanceType:                      &s.InstanceType,
			UserData:                          &userData,
			MaxCount:                          aws.Int64(1),
			MinCount:                          aws.Int64(1),
			IamInstanceProfile:                &ec2.IamInstanceProfileSpecification{Name: &s.IamInstanceProfile},
			BlockDeviceMappings:               s.BlockDevices.BuildLaunchDevices(),
			Placement:                         &ec2.Placement{AvailabilityZone: &s.AvailabilityZone},
			EbsOptimized:                      &s.EbsOptimized,
		}

		if s.SubnetId != "" && s.AssociatePublicIpAddress {
			runOpts.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
				&ec2.InstanceNetworkInterfaceSpecification{
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

		if s.ExpectedRootDevice == "ebs" {
			runOpts.InstanceInitiatedShutdownBehavior = &s.InstanceInitiatedShutdownBehavior
		}

		runResp, err := ec2conn.RunInstances(runOpts)
		if err != nil {
			err := fmt.Errorf("Error launching source instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		instanceId = *runResp.Instances[0].InstanceId
	} else {
		ui.Message(fmt.Sprintf(
			"Requesting spot instance '%s' for: %s",
			s.InstanceType, spotPrice))
		runSpotResp, err := ec2conn.RequestSpotInstances(&ec2.RequestSpotInstancesInput{
			SpotPrice: &spotPrice,
			LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
				KeyName:            &keyName,
				ImageId:            &s.SourceAMI,
				InstanceType:       &s.InstanceType,
				UserData:           &userData,
				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Name: &s.IamInstanceProfile},
				NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
					&ec2.InstanceNetworkInterfaceSpecification{
						DeviceIndex:              aws.Int64(0),
						AssociatePublicIpAddress: &s.AssociatePublicIpAddress,
						SubnetId:                 &s.SubnetId,
						Groups:                   securityGroupIds,
						DeleteOnTermination:      aws.Bool(true),
					},
				},
				Placement: &ec2.SpotPlacement{
					AvailabilityZone: &availabilityZone,
				},
				BlockDeviceMappings: s.BlockDevices.BuildLaunchDevices(),
				EbsOptimized:        &s.EbsOptimized,
			},
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
		stateChange := StateChangeConf{
			Pending:   []string{"open"},
			Target:    "active",
			Refresh:   SpotRequestStateRefreshFunc(ec2conn, *spotRequestId),
			StepState: state,
		}
		_, err = WaitForState(&stateChange)
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
	}

	// Set the instance ID so that the cleanup works properly
	s.instanceId = instanceId

	ui.Message(fmt.Sprintf("Instance ID: %s", instanceId))
	ui.Say(fmt.Sprintf("Waiting for instance (%v) to become ready...", instanceId))
	stateChange := StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "running",
		Refresh:   InstanceStateRefreshFunc(ec2conn, instanceId),
		StepState: state,
	}
	latestInstance, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance (%s) to become ready: %s", instanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instance := latestInstance.(*ec2.Instance)

	ec2Tags := make([]*ec2.Tag, 1, len(s.Tags)+1)
	ec2Tags[0] = &ec2.Tag{Key: aws.String("Name"), Value: aws.String("Packer Builder")}
	for k, v := range s.Tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
		Tags:      ec2Tags,
		Resources: []*string{instance.InstanceId},
	})
	if err != nil {
		ui.Message(
			fmt.Sprintf("Failed to tag a Name on the builder instance: %s", err))
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

func (s *StepRunSourceInstance) Cleanup(state multistep.StateBag) {

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
		stateChange := StateChangeConf{
			Pending: []string{"active", "open"},
			Refresh: SpotRequestStateRefreshFunc(ec2conn, *s.spotRequest.SpotInstanceRequestId),
			Target:  "cancelled",
		}

		WaitForState(&stateChange)

	}

	// Terminate the source instance if it exists
	if s.instanceId != "" {
		ui.Say("Terminating the source AWS instance...")
		if _, err := ec2conn.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: []*string{&s.instanceId}}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
			return
		}
		stateChange := StateChangeConf{
			Pending: []string{"pending", "running", "shutting-down", "stopped", "stopping"},
			Refresh: InstanceStateRefreshFunc(ec2conn, s.instanceId),
			Target:  "terminated",
		}

		WaitForState(&stateChange)
	}
}
