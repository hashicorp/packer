// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"
)

func TestCanceledShouldImmediatelyStopPolling(t *testing.T) {
	var testSubject = NewDeploymentPoller(func() (string, error) { return "Canceled", nil })
	testSubject.pause = func() { t.Fatal("Did not expect this to be called!") }

	res, err := testSubject.PollAsNeeded()
	if err != nil {
		t.Errorf("Expected PollAsNeeded to not return an error, but got '%s'.", err)
	}

	if res != "Canceled" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'Canceled', but got '%s' instead.", res)
	}
}

func TestFailedShouldImmediatelyStopPolling(t *testing.T) {
	var testSubject = NewDeploymentPoller(func() (string, error) { return "Failed", nil })
	testSubject.pause = func() { t.Fatal("Did not expect this to be called!") }

	res, err := testSubject.PollAsNeeded()
	if err != nil {
		t.Fatalf("Expected PollAsNeeded to not return an error, but got '%s'.", err)
	}

	if res != "Failed" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'Failed', but got '%s' instead.", res)
	}
}

func TestDeletedShouldImmediatelyStopPolling(t *testing.T) {
	var testSubject = NewDeploymentPoller(func() (string, error) { return "Deleted", nil })
	testSubject.pause = func() { t.Fatal("Did not expect this to be called!") }

	res, err := testSubject.PollAsNeeded()
	if err != nil {
		t.Fatalf("Expected PollAsNeeded to not return an error, but got '%s'.", err)
	}

	if res != "Deleted" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'Deleted', but got '%s' instead.", res)
	}
}

func TestSucceededShouldImmediatelyStopPolling(t *testing.T) {
	var testSubject = NewDeploymentPoller(func() (string, error) { return "Succeeded", nil })
	testSubject.pause = func() { t.Fatal("Did not expect this to be called!") }

	res, err := testSubject.PollAsNeeded()
	if err != nil {
		t.Fatalf("Expected PollAsNeeded to not return an error, but got '%s'.", err)
	}

	if res != "Succeeded" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'Succeeded', but got '%s' instead.", res)
	}
}

func TestPollerShouldPollOnNonStoppingStatus(t *testing.T) {
	count := 0

	var testSubject = NewDeploymentPoller(func() (string, error) { return "Succeeded", nil })
	testSubject.pause = func() { count += 1 }
	testSubject.getProvisioningState = func() (string, error) {
		count += 1
		switch count {
		case 0, 1:
			return "Working", nil
		default:
			return "Succeeded", nil
		}
	}

	res, err := testSubject.PollAsNeeded()
	if err != nil {
		t.Fatalf("Expected PollAsNeeded to not return an error, but got '%s'.", err)
	}

	if res != "Succeeded" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'Succeeded', but got '%s' instead.", res)
	}

	if count != 3 {
		t.Fatal("Expected DeploymentPoller to poll until 'Succeeded', but it did not.")
	}
}

func TestPollerShouldReturnErrorImmediately(t *testing.T) {
	var testSubject = NewDeploymentPoller(func() (string, error) { return "bad-bad-bad", fmt.Errorf("BOOM") })
	testSubject.pause = func() { t.Fatal("Did not expect this to be called!") }

	res, err := testSubject.PollAsNeeded()
	if err == nil {
		t.Fatal("Expected PollAsNeeded to return an error, but it did not.")
	}

	if res != "bad-bad-bad" {
		t.Fatalf("Expected PollAsNeeded to return a result of 'bad-bad-bad', but got '%s' instead.", res)
	}
}
