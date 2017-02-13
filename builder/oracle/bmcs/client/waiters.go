// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	"fmt"
	"time"
)

const (
	defaultWaitDurationMS = 5000
	defaultMaxRetries     = 0
)

type Waiter struct {
	WaitDurationMS int
	MaxRetries     int
}

type WaitableService interface {
	GetResourceState(id string) (string, error)
}

func stringSliceContains(slice []string, value string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}
	return false
}

// NewWaiter creates a waiter with default wait duration and unlimited retry
// operations.
func NewWaiter() *Waiter {
	return &Waiter{WaitDurationMS: defaultWaitDurationMS, MaxRetries: defaultMaxRetries}
}

// WaitForResourceToReachState polls a resource that implements WaitableService
// repeatedly until it reaches a known state or fails if it reaches an
// unexpected state. The duration of the interval and number of polls is
// determined by the Waiter configuration.
func (w *Waiter) WaitForResourceToReachState(svc WaitableService, id string, waitStates []string, terminalState string) error {
	for i := 0; w.MaxRetries == 0 || i < w.MaxRetries; i++ {
		state, err := svc.GetResourceState(id)
		if err != nil {
			return err
		}

		if stringSliceContains(waitStates, state) {
			time.Sleep(time.Duration(w.WaitDurationMS) * time.Millisecond)
			continue
		} else if state == terminalState {
			return nil
		}

		return fmt.Errorf("Unexpected resource state %s, expecting a waiting state %s or terminal state  %s ", state, waitStates, terminalState)
	}

	return fmt.Errorf("Maximum number of retries (%d) exceeded; resource did not reach state %s", w.MaxRetries, terminalState)
}
