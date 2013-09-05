package digitalocean

import (
	"fmt"
	"log"
	"time"
)

// waitForState simply blocks until the droplet is in
// a state we expect, while eventually timing out.
func waitForDropletState(desiredState string, dropletId uint, client *DigitalOceanClient, c config) error {
	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking droplet status... (attempt: %d)", attempts)
			_, status, err := client.DropletStatus(dropletId)
			if err != nil {
				result <- err
				return
			}

			if status == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)
		}
	}()

	log.Printf("Waiting for up to %s for droplet to become %s", c.RawStateTimeout, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(c.stateTimeout):
		err := fmt.Errorf("Timeout while waiting to for droplet to become '%s'", desiredState)
		return err
	}
}
