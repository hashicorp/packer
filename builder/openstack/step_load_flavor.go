package openstack

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/flavors"
)

// StepLoadFlavor gets the FlavorRef from a Flavor. It first assumes
// that the Flavor is a ref and verifies it. Otherwise, it tries to find
// the flavor by name.
type StepLoadFlavor struct {
	Flavor string
}

func (s *StepLoadFlavor) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Loading flavor: %s", s.Flavor))
	log.Printf("[INFO] Loading flavor by ID: %s", s.Flavor)
	flavor, err := flavors.Get(client, s.Flavor).Extract()
	if err != nil {
		log.Printf("[ERROR] Failed to find flavor by ID: %s", err)
		geterr := err

		log.Printf("[INFO] Loading flavor by name: %s", s.Flavor)
		id, err := flavors.IDFromName(client, s.Flavor)
		if err != nil {
			log.Printf("[ERROR] Failed to find flavor by name: %s", err)
			err = fmt.Errorf(
				"Unable to find specified flavor by ID or name!\n\n"+
					"Error from ID lookup: %s\n\n"+
					"Error from name lookup: %s",
				geterr,
				err)
			state.Put("error", err)
			return multistep.ActionHalt
		}

		flavor = &flavors.Flavor{ID: id}
	}

	ui.Message(fmt.Sprintf("Verified flavor. ID: %s", flavor.ID))
	state.Put("flavor_id", flavor.ID)
	return multistep.ActionContinue
}

func (s *StepLoadFlavor) Cleanup(state multistep.StateBag) {
}
