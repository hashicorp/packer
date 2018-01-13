package classic

import (
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	ui.Say("Getting Instance Info...")
	ui := state.Get("ui").(packer.Ui)
	instanceID := state.Get("instance_id").(string)
	endpoint_path := "/instance/%s", instanceID // GET

	// https://docs.oracle.com/en/cloud/iaas/compute-iaas-cloud/stcsa/op-instance-%7Bname%7D-get.html

}
