package googlecompute

import (
	"time"
)

// waitForInstanceState.
func waitForInstanceState(desiredState string, zone string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	return nil
	/*
		f := func() (string, error) {
			return client.InstanceStatus(zone, name)
		}
		return waitForState("instance", desiredState, f, timeout)
	*/
}

// waitForZoneOperationState.
func waitForZoneOperationState(desiredState string, zone string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	return nil
	/*
		f := func() (string, error) {
			return client.ZoneOperationStatus(zone, name)
		}
		return waitForState("operation", desiredState, f, timeout)
	*/
}

// waitForGlobalOperationState.
func waitForGlobalOperationState(desiredState string, name string, client *GoogleComputeClient, timeout time.Duration) error {
	/*
		f := func() (string, error) {
			return client.GlobalOperationStatus(name)
		}
		return waitForState("operation", desiredState, f, timeout)
	*/
	return nil
}
