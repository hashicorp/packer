package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepRunSourceInstance struct {
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
	SubnetId                          string
	Tags                              map[string]string
	VolumeTags                        map[string]string
	UserData                          string
	UserDataFile                      string
	Ctx                               interpolate.Context

	instanceId string
}

func (s *StepRunSourceInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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

	ec2Tags, err := ConvertToEC2Tags(s.Tags, *ec2conn.Config.Region, s.SourceAMI, s.Ctx)
	if err != nil {
		err := fmt.Errorf("Error tagging source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ReportTags(ui, ec2Tags)

	volTags, err := ConvertToEC2Tags(s.VolumeTags, *ec2conn.Config.Region, s.SourceAMI, s.Ctx)
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

	if len(tagSpecs) > 0 {
		runOpts.SetTagSpecifications(tagSpecs)
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
	if err := ec2conn.WaitUntilInstanceRunning(describeInstance); err != nil {
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
		stateChange := StateChangeConf{
			Pending: []string{"pending", "running", "shutting-down", "stopped", "stopping"},
			Refresh: InstanceStateRefreshFunc(ec2conn, s.instanceId),
			Target:  "terminated",
		}

		_, err := WaitForState(&stateChange)
		if err != nil {
			ui.Error(err.Error())
		}
	}
}
