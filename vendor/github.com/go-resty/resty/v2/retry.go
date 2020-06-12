// Copyright (c) 2015-2020 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"context"
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxRetries  = 3
	defaultWaitTime    = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

type (
	// Option is to create convenient retry options like wait time, max retries, etc.
	Option func(*Options)

	// RetryConditionFunc type is for retry condition function
	// input: non-nil Response OR request execution error
	RetryConditionFunc func(*Response, error) bool

	// RetryAfterFunc returns time to wait before retry
	// For example, it can parse HTTP Retry-After header
	// https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
	// Non-nil error is returned if it is found that request is not retryable
	// (0, nil) is a special result means 'use default algorithm'
	RetryAfterFunc func(*Client, *Response) (time.Duration, error)

	// Options struct is used to hold retry settings.
	Options struct {
		maxRetries      int
		waitTime        time.Duration
		maxWaitTime     time.Duration
		retryConditions []RetryConditionFunc
	}
)

// Retries sets the max number of retries
func Retries(value int) Option {
	return func(o *Options) {
		o.maxRetries = value
	}
}

// WaitTime sets the default wait time to sleep between requests
func WaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.waitTime = value
	}
}

// MaxWaitTime sets the max wait time to sleep between requests
func MaxWaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.maxWaitTime = value
	}
}

// RetryConditions sets the conditions that will be checked for retry.
func RetryConditions(conditions []RetryConditionFunc) Option {
	return func(o *Options) {
		o.retryConditions = conditions
	}
}

// Backoff retries with increasing timeout duration up until X amount of retries
// (Default is 3 attempts, Override with option Retries(n))
func Backoff(operation func() (*Response, error), options ...Option) error {
	// Defaults
	opts := Options{
		maxRetries:      defaultMaxRetries,
		waitTime:        defaultWaitTime,
		maxWaitTime:     defaultMaxWaitTime,
		retryConditions: []RetryConditionFunc{},
	}

	for _, o := range options {
		o(&opts)
	}

	var (
		resp *Response
		err  error
	)

	for attempt := 0; attempt <= opts.maxRetries; attempt++ {
		resp, err = operation()
		ctx := context.Background()
		if resp != nil && resp.Request.ctx != nil {
			ctx = resp.Request.ctx
		}
		if ctx.Err() != nil {
			return err
		}

		err1 := unwrapNoRetryErr(err)           // raw error, it used for return users callback.
		needsRetry := err != nil && err == err1 // retry on a few operation errors by default

		for _, condition := range opts.retryConditions {
			needsRetry = condition(resp, err1)
			if needsRetry {
				break
			}
		}

		if !needsRetry {
			return err
		}

		waitTime, err2 := sleepDuration(resp, opts.waitTime, opts.maxWaitTime, attempt)
		if err2 != nil {
			if err == nil {
				err = err2
			}
			return err
		}

		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

func sleepDuration(resp *Response, min, max time.Duration, attempt int) (time.Duration, error) {
	const maxInt = 1<<31 - 1 // max int for arch 386

	if max < 0 {
		max = maxInt
	}

	if resp == nil {
		goto defaultCase
	}

	// 1. Check for custom callback
	if retryAfterFunc := resp.Request.client.RetryAfter; retryAfterFunc != nil {
		result, err := retryAfterFunc(resp.Request.client, resp)
		if err != nil {
			return 0, err // i.e. 'API quota exceeded'
		}
		if result == 0 {
			goto defaultCase
		}
		if result < 0 || max < result {
			result = max
		}
		if result < min {
			result = min
		}
		return result, nil
	}

	// 2. Return capped exponential backoff with jitter
	// http://www.awsarchitectureblog.com/2015/03/backoff.html
defaultCase:
	base := float64(min)
	capLevel := float64(max)

	temp := math.Min(capLevel, base*math.Exp2(float64(attempt)))
	ri := int(temp / 2)
	if ri <= 0 {
		ri = maxInt // max int for arch 386
	}
	result := time.Duration(math.Abs(float64(ri + rand.Intn(ri))))

	if result < min {
		result = min
	}

	return result, nil
}
