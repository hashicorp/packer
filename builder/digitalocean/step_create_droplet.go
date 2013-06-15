package digitalocean

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

type stepCreateDroplet struct {
	dropletId uint
}

func (s *stepCreateDroplet) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	c := state["config"].(config)
	sshKeyId := state["ssh_key_id"].(uint)

	ui.Say("Creating droplet...")

	// Some random droplet name as it's temporary
	name := fmt.Sprintf("packer-%s", hex.EncodeToString(identifier.NewUUID().Raw()))

	// Create the droplet based on configuration
	dropletId, err := client.CreateDroplet(name, c.SizeID, c.ImageID, c.RegionID, sshKeyId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.dropletId = dropletId

	// Store the droplet id for later
	state["droplet_id"] = dropletId

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) Cleanup(state map[string]interface{}) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")

	// Sleep arbitrarily before sending destroy request
	// Otherwise we get "pending event" errors, even though there isn't
	// one.
	time.Sleep(5 * time.Second)

	err := client.DestroyDroplet(s.dropletId)

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %v", s.dropletId))
	}
}
