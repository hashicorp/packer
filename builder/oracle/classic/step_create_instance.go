package classic

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateInstance struct{}

func (s *stepCreateIPReservation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Instance...")
	const endpoint_path = "/launchplan/" // POST
	// master-instance.json
	`
	{
	  "instances": [{
	      "shape": "oc3",
	      "sshkeys": ["/Compute-mydomain/user@example.com/my_sshkey"],
	      "name": "Compute-mydomain/user@example.com/master-instance",
	      "label": "master-instance",
	      "imagelist": "/Compute-mydomain/user@example.com/Ubuntu.16.04-LTS.amd64.20170330",
	      "networking": {
	        "eth0": {
	          "nat": "ipreservation:/Compute-mydomain/user@example.com/master-instance-ip"
	        }
	      }
	  }]
	}
	`
	// command line call
	// $ opc compute launch-plans add --request-body=./master-instance.json
	// ...

	instanceID, err := client.CreateInstance(publicKey)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_id", instanceID)

	ui.Say(fmt.Sprintf("Created instance (%s).", instanceID))
}
