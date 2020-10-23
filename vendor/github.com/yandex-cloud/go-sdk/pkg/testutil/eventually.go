// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package testutil

import (
	"context"
	"time"

	testing "github.com/mitchellh/go-testing-interface"
)

func Eventually(t testing.T, check CheckFunc, opts ...EventuallyOption) bool {
	options := &eventuallyOptions{
		timeout:      5 * time.Second,
		pollInterval: 10 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.timeout)
	defer cancel()

	success := make(chan struct{})
	// check could block, so it should be done in a separate goroutine.
	go func() {
		ticker := time.NewTicker(options.pollInterval)
		defer ticker.Stop()
		for {
			ok := check()
			if ok {
				close(success)
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("Eventually condition not met: "+options.messageFormat, options.formatArgs...)
		return false
	case <-success:
		return true
	}
}

type eventuallyOptions struct {
	timeout      time.Duration
	pollInterval time.Duration

	messageFormat string
	formatArgs    []interface{}
}

func PollTimeout(timeout time.Duration) EventuallyOption {
	return func(opts *eventuallyOptions) {
		opts.timeout = timeout
	}
}

func PollInterval(interval time.Duration) EventuallyOption {
	return func(opts *eventuallyOptions) {
		opts.pollInterval = interval
	}
}

type CheckFunc func() bool

type EventuallyOption func(opts *eventuallyOptions)

func Message(format string, args ...interface{}) EventuallyOption {
	return func(opts *eventuallyOptions) {
		opts.messageFormat = format
		opts.formatArgs = args
	}
}
