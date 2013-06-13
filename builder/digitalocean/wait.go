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
			select {
			default:
			}

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

			// Wait a second in between
			time.Sleep(1 * time.Second)
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
		case <-time.After(1 * time.Second):
			err := errors.New("Interrupt detected, quitting waiting for droplet")
			return err
		}
	}

	// If we got this far, there were no errors
	return nil
}
