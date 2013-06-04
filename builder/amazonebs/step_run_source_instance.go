package amazonebs

import (
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepRunSourceInstance struct {
	instance *ec2.Instance
}

func (s *stepRunSourceInstance) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ec2conn := state["ec2"].(*ec2.EC2)
	keyName := state["keyPair"].(string)
	ui := state["ui"].(packer.Ui)

	runOpts := &ec2.RunInstances{
		KeyName:      keyName,
		ImageId:      config.SourceAmi,
		InstanceType: config.InstanceType,
		MinCount:     0,
		MaxCount:     0,
	}

	ui.Say("Launching a source AWS instance...")
	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.instance = &runResp.Instances[0]
	log.Printf("instance id: %s", s.instance.InstanceId)

	ui.Say("Waiting for instance to become ready...")
	s.instance, err = waitForState(ec2conn, s.instance, []string{"pending"}, "running")
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["instance"] = s.instance

	return multistep.ActionContinue
}

func (s *stepRunSourceInstance) Cleanup(state map[string]interface{}) {
	if s.instance == nil {
		return
	}

	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	// TODO(mitchellh): error handling
	ui.Say("Terminating the source AWS instance...")
	ec2conn.TerminateInstances([]string{s.instance.InstanceId})
}
