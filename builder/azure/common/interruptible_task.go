// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package common

import (
	"time"
)

type InterruptibleTaskResult struct {
	Err         error
	IsCancelled bool
}

type InterruptibleTask struct {
	IsCancelled func() bool
	Task        func(cancelCh <-chan struct{}) error
}

func NewInterruptibleTask(isCancelled func() bool, task func(cancelCh <-chan struct{}) error) *InterruptibleTask {
	return &InterruptibleTask{
		IsCancelled: isCancelled,
		Task:        task,
	}
}

func StartInterruptibleTask(isCancelled func() bool, task func(cancelCh <-chan struct{}) error) InterruptibleTaskResult {
	t := NewInterruptibleTask(isCancelled, task)
	return t.Run()
}

func (s *InterruptibleTask) Run() InterruptibleTaskResult {
	completeCh := make(chan error)

	cancelCh := make(chan struct{})
	defer close(cancelCh)

	go func() {
		err := s.Task(cancelCh)
		completeCh <- err

		// senders close, receivers check for close
		close(completeCh)
	}()

	for {
		if s.IsCancelled() {
			return InterruptibleTaskResult{Err: nil, IsCancelled: true}
		}

		select {
		case err := <-completeCh:
			return InterruptibleTaskResult{Err: err, IsCancelled: false}
		case <-time.After(100 * time.Millisecond):
		}
	}
}
