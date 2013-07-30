package chroot

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepCheckEC2 verifies that this builder is running on an EC2 instance.
type StepCheckEC2 struct{}

func (s *StepCheckEC2) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)

	ui.Say("Verifying we're on an EC2 instance...")
	id, err := aws.GetMetaData("instance-id")
	if err != nil {
		log.Printf("Error: %s", err)
		err := fmt.Errorf(
			"Error retrieving the ID of the instance Packer is running on.\n" +
				"Please verify Packer is running on a proper AWS EC2 instance.")
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Instance ID: %s", string(id))

	return multistep.ActionContinue
}

func (s *StepCheckEC2) Cleanup(map[string]interface{}) {}
