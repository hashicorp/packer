package digitalocean

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepGetDefaults struct{}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func (s *stepGetDefaults) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)
	client := state.Get("client").(DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Get default regio, size and image for droplet...")

	regions, err := client.Regions()
	if err != nil {
		err := fmt.Errorf("Error getting regions: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defaultRegion := Region{}
	defaultRegion = regions[random(0, len(regions))]
	if defaultRegion.Slug == "" {
		err := fmt.Errorf("Error getting default region: %s", "get empty region slug")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	config.Region = defaultRegion.Slug

	sizes, err := client.Sizes()
	if err != nil {
		err := fmt.Errorf("Error getting sizes: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defaultSize := Size{}
	for _, size := range sizes {
		for _, region := range size.Regions {
			if region == defaultRegion.Slug {
				if defaultSize.Slug == "" {
					defaultSize = size
				} else {
					if defaultSize.Memory > size.Memory {
						defaultSize = size
					}
				}
			}
		}
	}
	if defaultSize.Slug == "" {
		err := fmt.Errorf("Error getting default size: %s", "get empty size slug")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	config.Size = defaultSize.Slug

	return multistep.ActionContinue
}

func (s *stepGetDefaults) Cleanup(state multistep.StateBag) {}
