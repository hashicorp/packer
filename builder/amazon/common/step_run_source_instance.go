package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type StepRunSourceInstance struct {
	ExpectedRootDevice string
	InstanceType       string
	UserData           string
	SourceAMI          string
	IamInstanceProfile string
	SubnetId           string

	instance *ec2.Instance
}

func (s *StepRunSourceInstance) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	keyName := state["keyPair"].(string)
	securityGroupId := state["securityGroupId"].(string)
	ui := state["ui"].(packer.Ui)

	runOpts := &ec2.RunInstances{
		KeyName:            keyName,
		ImageId:            s.SourceAMI,
		InstanceType:       s.InstanceType,
		UserData:           []byte(s.UserData),
		MinCount:           0,
		MaxCount:           0,
		SecurityGroups:     []ec2.SecurityGroup{ec2.SecurityGroup{Id: securityGroupId}},
		IamInstanceProfile: s.IamInstanceProfile,
		SubnetId:           s.SubnetId,
	}

	ui.Say("Launching a source AWS instance...")
	imageResp, err := ec2conn.Images([]string{s.SourceAMI}, ec2.NewFilter())
	if err != nil {
		state["error"] = fmt.Errorf("There was a problem with the source AMI: %s", err)
		return multistep.ActionHalt
	}

	if len(imageResp.Images) != 1 {
		state["error"] = fmt.Errorf("The source AMI '%s' could not be found.", s.SourceAMI)
		return multistep.ActionHalt
	}

	if s.ExpectedRootDevice != "" && imageResp.Images[0].RootDeviceType != s.ExpectedRootDevice {
		state["error"] = fmt.Errorf(
			"The provided source AMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, imageResp.Images[0].RootDeviceType)
		return multistep.ActionHalt
	}

	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		err := fmt.Errorf("Error launching source instance: %s", err)
		state["error"] = err
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
	s.instance = latestInstance.(*ec2.Instance)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance (%s) to become ready: %s", s.instance.InstanceId, err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["instance"] = s.instance

	return multistep.ActionContinue
}

func (s *StepRunSourceInstance) Cleanup(state map[string]interface{}) {
	if s.instance == nil {
		return
	}

	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

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
