package openstack

import (
	"context"
	"fmt"
	"log"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepLoadExtensions gets the FlavorRef from a Flavor. It first assumes
// that the Flavor is a ref and verifies it. Otherwise, it tries to find
// the flavor by name.
type StepLoadExtensions struct{}

func (s *StepLoadExtensions) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Discovering enabled extensions...")
	result := make(map[string]struct{}, 15)
	pager := extensions.List(client)
	err = pager.EachPage(func(p pagination.Page) (bool, error) {
		// Extract the extensions from this page
		exts, err := extensions.ExtractExtensions(p)
		if err != nil {
			return false, err
		}

		for _, ext := range exts {
			log.Printf("[DEBUG] Discovered extension: %s", ext.Alias)
			result[ext.Alias] = struct{}{}
		}

		return true, nil
	})
	if err != nil {
		err = fmt.Errorf("Error loading extensions: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("extensions", result)
	return multistep.ActionContinue
}

func (s *StepLoadExtensions) Cleanup(state multistep.StateBag) {
}
