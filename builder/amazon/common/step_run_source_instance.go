package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
)

type StepRunSourceInstance struct {
	AssociatePublicIpAddress bool
	SpotPrice                string
	AvailabilityZone         string
	BlockDevices             BlockDevices
	Debug                    bool
	ExpectedRootDevice       string
	InstanceType             string
	IamInstanceProfile       string
	SourceAMI                string
	SubnetId                 string
	Tags                     map[string]string
	UserData                 string
	UserDataFile             string
	spotRequest              *ec2.SpotRequestResult
	instance                 *ec2.Instance
}

func (s *StepRunSourceInstance) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	keyName := state.Get("keyPair").(string)
	securityGroupIds := state.Get("securityGroupIds").([]string)
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

	securityGroups := make([]ec2.SecurityGroup, len(securityGroupIds))
	for n, securityGroupId := range securityGroupIds {
		securityGroups[n] = ec2.SecurityGroup{Id: securityGroupId}
	}

	ui.Say("Launching a source AWS instance...")
	imageResp, err := ec2conn.Images([]string{s.SourceAMI}, ec2.NewFilter())
	if err != nil {
		state.Put("error", fmt.Errorf("There was a problem with the source AMI: %s", err))
		return multistep.ActionHalt
	}

	if len(imageResp.Images) != 1 {
		state.Put("error", fmt.Errorf("The source AMI '%s' could not be found.", s.SourceAMI))
		return multistep.ActionHalt
	}

	if s.ExpectedRootDevice != "" && imageResp.Images[0].RootDeviceType != s.ExpectedRootDevice {
		state.Put("error", fmt.Errorf(
			"The provided source AMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, imageResp.Images[0].RootDeviceType))
		return multistep.ActionHalt
	}

	var instanceId []string
	if s.SpotPrice == "" {
		runOpts := &ec2.RunInstances{
			KeyName:                  keyName,
			ImageId:                  s.SourceAMI,
			InstanceType:             s.InstanceType,
			UserData:                 []byte(userData),
			MinCount:                 0,
			MaxCount:                 0,
			SecurityGroups:           securityGroups,
			IamInstanceProfile:       s.IamInstanceProfile,
			SubnetId:                 s.SubnetId,
			AssociatePublicIpAddress: s.AssociatePublicIpAddress,
			BlockDevices:             s.BlockDevices.BuildLaunchDevices(),
			AvailZone:                s.AvailabilityZone,
		}
		runResp, err := ec2conn.RunInstances(runOpts)
		if err != nil {
			err := fmt.Errorf("Error launching source instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		instanceId = []string{runResp.Instances[0].InstanceId}
	} else {
		runOpts := &ec2.RequestSpotInstances{
			SpotPrice:                s.SpotPrice,
			KeyName:                  keyName,
			ImageId:                  s.SourceAMI,
			InstanceType:             s.InstanceType,
			UserData:                 []byte(userData),
			SecurityGroups:           securityGroups,
			IamInstanceProfile:       s.IamInstanceProfile,
			SubnetId:                 s.SubnetId,
			AssociatePublicIpAddress: s.AssociatePublicIpAddress,
			BlockDevices:             s.BlockDevices.BuildLaunchDevices(),
			AvailZone:                s.AvailabilityZone,
		}
		runSpotResp, err := ec2conn.RequestSpotInstances(runOpts)
		if err != nil {
			err := fmt.Errorf("Error launching source spot instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		spotRequestId := runSpotResp.SpotRequestResults[0].SpotRequestId
		ui.Say(fmt.Sprintf("Waiting for spot request (%s) to become ready...", spotRequestId))
		stateChange := StateChangeConf{
			Pending:   []string{"open"},
			Target:    "active",
			Refresh:   SpotRequestStateRefreshFunc(ec2conn, spotRequestId),
			StepState: state,
		}
		_, err = WaitForState(&stateChange)
		if err != nil {
			err := fmt.Errorf("Error waiting for spot request (%s) to become ready: %s", spotRequestId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		spotResp, err := ec2conn.DescribeSpotRequests([]string{spotRequestId}, nil)
		if err != nil {
			err := fmt.Errorf("Error finding spot request (%s): %s", spotRequestId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.spotRequest = &spotResp.SpotRequestResults[0]
		instanceId = []string{s.spotRequest.InstanceId}
	}

	instanceResp, err := ec2conn.Instances(instanceId, nil)
	if err != nil {
		err := fmt.Errorf("Error finding source instance (%s): %s", instanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.instance = &instanceResp.Reservations[0].Instances[0]
	ui.Message(fmt.Sprintf("Instance ID: %s", s.instance.InstanceId))

	ec2Tags := make([]ec2.Tag, 1, len(s.Tags)+1)
	ec2Tags[0] = ec2.Tag{"Name", "Packer Builder"}
	for k, v := range s.Tags {
		ec2Tags = append(ec2Tags, ec2.Tag{k, v})
	}

	_, err = ec2conn.CreateTags([]string{s.instance.InstanceId}, ec2Tags)
	if err != nil {
		ui.Message(
			fmt.Sprintf("Failed to tag a Name on the builder instance: %s", err))
	}

	ui.Say(fmt.Sprintf("Waiting for instance (%s) to become ready...", s.instance.InstanceId))
	stateChange := StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "running",
		Refresh:   InstanceStateRefreshFunc(ec2conn, s.instance),
		StepState: state,
	}
	latestInstance, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance (%s) to become ready: %s", s.instance.InstanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.instance = latestInstance.(*ec2.Instance)

	if s.Debug {
		if s.instance.DNSName != "" {
			ui.Message(fmt.Sprintf("Public DNS: %s", s.instance.DNSName))
		}

		if s.instance.PublicIpAddress != "" {
			ui.Message(fmt.Sprintf("Public IP: %s", s.instance.PublicIpAddress))
		}

		if s.instance.PrivateIpAddress != "" {
			ui.Message(fmt.Sprintf("Private IP: %s", s.instance.PrivateIpAddress))
		}
	}

	state.Put("instance", s.instance)

	return multistep.ActionContinue
}

func (s *StepRunSourceInstance) Cleanup(state multistep.StateBag) {

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// Cancel the spot request if it exists
	if s.spotRequest != nil {
		ui.Say("Cancelling the spot request...")
		if _, err := ec2conn.CancelSpotRequests([]string{s.spotRequest.SpotRequestId}); err != nil {
			ui.Error(fmt.Sprintf("Error cancelling the spot request, may still be around: %s", err))
			return
		}
		stateChange := StateChangeConf{
			Pending: []string{"active", "open"},
			Refresh: SpotRequestStateRefreshFunc(ec2conn, s.spotRequest.SpotRequestId),
			Target:  "cancelled",
		}

		WaitForState(&stateChange)

	}

	// Terminate the source instance if it exists
	if s.instance != nil {

		ui.Say("Terminating the source AWS instance...")
		if _, err := ec2conn.TerminateInstances([]string{s.instance.InstanceId}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
			return
		}
		stateChange := StateChangeConf{
			Pending: []string{"pending", "running", "shutting-down", "stopped", "stopping"},
			Refresh: InstanceStateRefreshFunc(ec2conn, s.instance),
			Target:  "terminated",
		}

		WaitForState(&stateChange)
	}
}
