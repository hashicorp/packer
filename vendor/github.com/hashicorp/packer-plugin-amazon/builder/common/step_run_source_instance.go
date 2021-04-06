package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/packer-plugin-amazon/builder/common/awserrors"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type StepRunSourceInstance struct {
	PollingConfig                     *AWSPollingConfig
	AssociatePublicIpAddress          bool
	LaunchMappings                    EC2BlockDeviceMappingsBuilder
	Comm                              *communicator.Config
	Ctx                               interpolate.Context
	Debug                             bool
	EbsOptimized                      bool
	EnableT2Unlimited                 bool
	ExpectedRootDevice                string
	HttpEndpoint                      string
	HttpTokens                        string
	HttpPutResponseHopLimit           int64
	InstanceInitiatedShutdownBehavior string
	InstanceType                      string
	IsRestricted                      bool
	SourceAMI                         string
	Tags                              map[string]string
	Tenancy                           string
	UserData                          string
	UserDataFile                      string
	VolumeTags                        map[string]string
	NoEphemeral                       bool

	instanceId string
}

func (s *StepRunSourceInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)

	securityGroupIds := aws.StringSlice(state.Get("securityGroupIds").([]string))
	iamInstanceProfile := aws.String(state.Get("iamInstanceProfile").(string))

	ui := state.Get("ui").(packersdk.Ui)

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

	ec2Tags, err := TagMap(s.Tags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	volTags, err := TagMap(s.VolumeTags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging volumes: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	az := state.Get("availability_zone").(string)
	runOpts := &ec2.RunInstancesInput{
		ImageId:             &s.SourceAMI,
		InstanceType:        &s.InstanceType,
		UserData:            &userData,
		MaxCount:            aws.Int64(1),
		MinCount:            aws.Int64(1),
		IamInstanceProfile:  &ec2.IamInstanceProfileSpecification{Name: iamInstanceProfile},
		BlockDeviceMappings: s.LaunchMappings.BuildEC2BlockDeviceMappings(),
		Placement:           &ec2.Placement{AvailabilityZone: &az},
		EbsOptimized:        &s.EbsOptimized,
	}

	if s.NoEphemeral {
		// This is only relevant for windows guests. Ephemeral drives by
		// default are assigned to drive names xvdca-xvdcz.
		// When vms are launched from the AWS console, they're automatically
		// removed from the block devices if the user hasn't said to use them,
		// but the SDK does not perform this cleanup. The following code just
		// manually removes the ephemeral drives from the mapping so that they
		// don't clutter up console views and cause confusion.
		log.Printf("no_ephemeral was set, so creating drives xvdca-xvdcz as empty mappings")
		DefaultEphemeralDeviceLetters := "abcdefghijklmnopqrstuvwxyz"
		for _, letter := range DefaultEphemeralDeviceLetters {
			bd := &ec2.BlockDeviceMapping{
				DeviceName: aws.String("xvdc" + string(letter)),
				NoDevice:   aws.String(""),
			}
			runOpts.BlockDeviceMappings = append(runOpts.BlockDeviceMappings, bd)
		}
	}

	if s.EnableT2Unlimited {
		creditOption := "unlimited"
		runOpts.CreditSpecification = &ec2.CreditSpecificationRequest{CpuCredits: &creditOption}
	}

	if s.HttpEndpoint == "enabled" {
		runOpts.MetadataOptions = &ec2.InstanceMetadataOptionsRequest{HttpEndpoint: &s.HttpEndpoint, HttpTokens: &s.HttpTokens, HttpPutResponseHopLimit: &s.HttpPutResponseHopLimit}
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

	if s.Comm.SSHKeyPairName != "" {
		runOpts.KeyName = &s.Comm.SSHKeyPairName
	}

	subnetId := state.Get("subnet_id").(string)

	if subnetId != "" && s.AssociatePublicIpAddress {
		runOpts.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: &s.AssociatePublicIpAddress,
				SubnetId:                 aws.String(subnetId),
				Groups:                   securityGroupIds,
				DeleteOnTermination:      aws.Bool(true),
			},
		}
	} else {
		runOpts.SubnetId = aws.String(subnetId)
		runOpts.SecurityGroupIds = securityGroupIds
	}

	if s.ExpectedRootDevice == "ebs" {
		runOpts.InstanceInitiatedShutdownBehavior = &s.InstanceInitiatedShutdownBehavior
	}

	if s.Tenancy != "" {
		runOpts.Placement.Tenancy = aws.String(s.Tenancy)
	}

	var runResp *ec2.Reservation
	err = retry.Config{
		Tries: 11,
		ShouldRetry: func(err error) bool {
			if awserrors.Matches(err, "InvalidParameterValue", "iamInstanceProfile") {
				return true
			}
			return false
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		runResp, err = ec2conn.RunInstances(runOpts)
		return err
	})

	if awserrors.Matches(err, "VPCIdNotSpecified", "No default VPC for this user") && subnetId == "" {
		err := fmt.Errorf("Error launching source instance: a valid Subnet Id was not specified")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

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

	if err := s.PollingConfig.WaitUntilInstanceRunning(ctx, ec2conn, instanceId); err != nil {
		err := fmt.Errorf("Error waiting for instance (%s) to become ready: %s", instanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())

		// try to get some context from AWS on why was instance
		// transitioned to the unexpected state
		if resp, e := ec2conn.DescribeInstances(describeInstance); e == nil {
			if len(resp.Reservations) > 0 && len(resp.Reservations[0].Instances) > 0 {
				instance := resp.Reservations[0].Instances[0]
				if instance.StateTransitionReason != nil && instance.StateReason != nil && instance.StateReason.Message != nil {
					ui.Error(fmt.Sprintf("Instance state change details: %s: %s",
						*instance.StateTransitionReason, *instance.StateReason.Message))
				}
			}
		}
		return multistep.ActionHalt
	}

	// there's a race condition that can happen because of AWS's eventual
	// consistency where even though the wait is complete, the describe call
	// will fail. Retry a couple of times to try to mitigate that race.

	var r *ec2.DescribeInstancesOutput
	err = retry.Config{Tries: 11, ShouldRetry: func(err error) bool {
		if awserrors.Matches(err, "InvalidInstanceID.NotFound", "") {
			return true
		}
		return false
	},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		r, err = ec2conn.DescribeInstances(describeInstance)
		return err
	})
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", instance.InstanceId)

	// If we're in a region that doesn't support tagging on instance creation,
	// do that now.

	if s.IsRestricted {
		ec2Tags.Report(ui)
		// Retry creating tags for about 2.5 minutes
		err = retry.Config{Tries: 11, ShouldRetry: func(error) bool {
			if awserrors.Matches(err, "InvalidInstanceID.NotFound", "") {
				return true
			}
			return false
		},
			RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			_, err := ec2conn.CreateTags(&ec2.CreateTagsInput{
				Tags:      ec2Tags,
				Resources: []*string{instance.InstanceId},
			})
			return err
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

		if len(volumeIds) > 0 && len(s.VolumeTags) > 0 {
			ui.Say("Adding tags to source EBS Volumes")

			volumeTags, err := TagMap(s.VolumeTags).EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
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
	ui := state.Get("ui").(packersdk.Ui)

	// Terminate the source instance if it exists
	if s.instanceId != "" {
		ui.Say("Terminating the source AWS instance...")
		if _, err := ec2conn.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: []*string{&s.instanceId}}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
			return
		}

		if err := s.PollingConfig.WaitUntilInstanceTerminated(aws.BackgroundContext(), ec2conn, s.instanceId); err != nil {
			ui.Error(err.Error())
		}
	}
}
