package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	c := state["config"].(config)
	dropletId := state["droplet_id"].(uint)

	ui.Say("Waiting for droplet to become active...")

	// Wait for the droplet to become active
	active := make(chan bool, 1)

	go func() {
		var err error

		attempts := 0
		for {
			select {
			default:
			}

			attempts += 1

			log.Printf("Checking droplet status... (attempt: %d)", attempts)

			ip, status, err := client.DropletStatus(dropletId)

			if status == "active" {
				break
			}

			// Wait a second in between
			time.Sleep(1 * time.Second)
		}

		active <- true
	}()

	log.Printf("Waiting for up to 3 minutes for droplet to become active")
	duration, _ := time.ParseDuration("3m")
	timeout := time.After(duration)

ActiveWaitLoop:
	for {
		select {
		case <-active:
			// We connected. Just break the loop.
			break ActiveWaitLoop
		case <-timeout:
			ui.Error("Timeout while waiting to for droplet to become active")
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				log.Println("Interrupt detected, quitting waiting droplet to become active")
				return multistep.ActionHalt
			}
		}
	}

	// Set the IP on the state for later
	ip, _, err := client.DropletStatus(dropletId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["droplet_ip"] = ip

	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state map[string]interface{}) {
	// no cleanup
}
