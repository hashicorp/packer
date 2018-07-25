package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	retry "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepRunSourceInstance struct {
	AssociatePublicIpAddress          bool
	AvailabilityZone                  string
	BlockDevices                      BlockDevices
	Ctx                               interpolate.Context
	Debug                             bool
	EbsOptimized                      bool
	EnableT2Unlimited                 bool
	ExpectedRootDevice                string
	IamInstanceProfile                string
	InstanceInitiatedShutdownBehavior string
	InstanceType                      string
	IsRestricted                      bool
	SourceAMI                         string
	SubnetId                          string
	Tags                              TagMap
	UserData                          string
	UserDataFile                      string
	VolumeTags                        TagMap

	instanceId string
}

func (s *StepRunSourceInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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

	volTags, err := s.VolumeTags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging volumes: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	runOpts := &ec2.RunInstancesInput{
		ImageId:             &s.SourceAMI,
		InstanceType:        &s.InstanceType,
		UserData:            &userData,
		MaxCount:            aws.Int64(1),
		MinCount:            aws.Int64(1),
		IamInstanceProfile:  &ec2.IamInstanceProfileSpecification{Name: &s.IamInstanceProfile},
		BlockDeviceMappings: s.BlockDevices.BuildLaunchDevices(),
		Placement:           &ec2.Placement{AvailabilityZone: &s.AvailabilityZone},
		EbsOptimized:        &s.EbsOptimized,
	}

	if s.EnableT2Unlimited {
		creditOption := "unlimited"
		runOpts.CreditSpecification = &ec2.CreditSpecificationRequest{CpuCredits: &creditOption}
	}

	// Collect tags for tagging on resource creation
	var tagSpecs []*ec2.TagSpecification

	if len(ec2Tags) > 0 {
		runTags := &ec2.TagSpecification{
			ResourceType: aws.String("instance"),
			Tags:         ec2Tags,
		}

		tagSpecs = append(tagSpecs, runTags)
	}

	if len(volTags) > 0 {
		runVolTags := &ec2.TagSpecification{
			ResourceType: aws.String("volume"),
			Tags:         volTags,
		}

		tagSpecs = append(tagSpecs, runVolTags)
	}

	// If our region supports it, set tag specifications
	if len(tagSpecs) > 0 && !s.IsRestricted {
		runOpts.SetTagSpecifications(tagSpecs)
		ec2Tags.Report(ui)
		volTags.Report(ui)
	}

	if keyName != "" {
		runOpts.KeyName = &keyName
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

	r, err := ec2conn.DescribeInstances(describeInstance)

	if err != nil || len(r.Reservations) == 0 || len(r.Reservations[0].Instances) == 0 {
		err := fmt.Errorf("Error finding source instance.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instance := r.Reservations[0].Instances[0]

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

	// If we're in a region that doesn't support tagging on instance creation,
	// do that now.

	if s.IsRestricted {
		ec2Tags.Report(ui)
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

		// Now tag volumes

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
	}

	return multistep.ActionContinue
}

func (s *StepRunSourceInstance) Cleanup(state multistep.StateBag) {

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

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
