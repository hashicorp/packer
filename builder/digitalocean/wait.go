package digitalocean

import (
	"errors"
	"log"
	"time"
)

// waitForState simply blocks until the droplet is in
// a state we expect, while eventually timing out.
func waitForDropletState(desiredState string, dropletId uint, client *DigitalOceanClient) error {
	active := make(chan bool, 1)

	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking droplet status... (attempt: %d)", attempts)

			_, status, err := client.DropletStatus(dropletId)

			if err != nil {
				log.Println(err)
				break
			}

			if status == desiredState {
				break
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)
		}

		active <- true
	}()

	log.Printf("Waiting for up to 3 minutes for droplet to become %s", desiredState)
	duration, _ := time.ParseDuration("3m")
	timeout := time.After(duration)

ActiveWaitLoop:
	for {
		select {
		case <-active:
			// We connected. Just break the loop.
			break ActiveWaitLoop
		case <-timeout:
			err := errors.New("Timeout while waiting to for droplet to become active")
			return err
		}
	}

	// If we got this far, there were no errors
	return nil
}
