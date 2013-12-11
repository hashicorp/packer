package googlecompute

import (
	"fmt"
	"log"
	"time"
)

// statusFunc.
type statusFunc func() (string, error)

// waitForInstanceState.
func waitForInstanceState(desiredState string, zone string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	f := func() (string, error) {
		return client.InstanceStatus(zone, name)
	}
	return waitForState("instance", desiredState, f, timeout)
}

// waitForZoneOperationState.
func waitForZoneOperationState(desiredState string, zone string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	f := func() (string, error) {
		return client.ZoneOperationStatus(zone, name)
	}
	return waitForState("operation", desiredState, f, timeout)
}

// waitForGlobalOperationState.
func waitForGlobalOperationState(desiredState string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	f := func() (string, error) {
		return client.GlobalOperationStatus(name)
	}
	return waitForState("operation", desiredState, f, timeout)
}

// waitForState.
func waitForState(kind string, desiredState string, f statusFunc, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)
	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1
			log.Printf("Checking %s state... (attempt: %d)", kind, attempts)
			status, err := f()
			if err != nil {
				result <- err
				return
			}
			if status == desiredState {
				result <- nil
				return
			}
			time.Sleep(3 * time.Second)
			select {
			case <-done:
				return
			default:
				continue
			}
		}
	}()
	log.Printf("Waiting for up to %d seconds for %s to become %s", timeout, kind, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for the %s to become '%s'", kind, desiredState)
		return err
	}
}
