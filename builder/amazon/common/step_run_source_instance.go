package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
)

type StepRunSourceInstance struct {
	Debug              bool
	ExpectedRootDevice string
	InstanceType       string
	UserData           string
	UserDataFile       string
	SourceAMI          string
	IamInstanceProfile string
	SubnetId           string
	AvailabilityZone   string
	BlockDevices       BlockDevices

	instance *ec2.Instance
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

	runOpts := &ec2.RunInstances{
		KeyName:            keyName,
		ImageId:            s.SourceAMI,
		InstanceType:       s.InstanceType,
		UserData:           []byte(userData),
		MinCount:           0,
		MaxCount:           0,
		SecurityGroups:     securityGroups,
		IamInstanceProfile: s.IamInstanceProfile,
		SubnetId:           s.SubnetId,
		BlockDevices:       s.BlockDevices.BuildLaunchDevices(),
		AvailZone:          s.AvailabilityZone,
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

	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		err := fmt.Errorf("Error launching source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.instance = &runResp.Instances[0]
	log.Printf("instance id: %s", s.instance.InstanceId)

	ui.Say(fmt.Sprintf("Waiting for instance (%s) to become ready...", s.instance.InstanceId))
	stateChange := StateChangeConf{
		Conn:      ec2conn,
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

		if s.instance.PrivateIpAddress != "" {
			ui.Message(fmt.Sprintf("Private IP: %s", s.instance.PrivateIpAddress))
		}
	}

	state.Put("instance", s.instance)

	return multistep.ActionContinue
}

func (s *StepRunSourceInstance) Cleanup(state multistep.StateBag) {
	if s.instance == nil {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Terminating the source AWS instance...")
	if _, err := ec2conn.TerminateInstances([]string{s.instance.InstanceId}); err != nil {
		ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Conn:    ec2conn,
		Pending: []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Refresh: InstanceStateRefreshFunc(ec2conn, s.instance),
		Target:  "terminated",
	}

	WaitForState(&stateChange)
}
