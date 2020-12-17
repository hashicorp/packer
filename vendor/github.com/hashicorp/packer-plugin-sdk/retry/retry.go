// Package retry provides tooling to retry API calls which are known to be
// vulnerable to throttling or flakiness due to eventual consistency.
package retry

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Config represents a retry config
type Config struct {
	// The operation will be retried until StartTimeout has elapsed. 0 means
	// forever.
	StartTimeout time.Duration

	// RetryDelay gives the time elapsed after a failure and before we try
	// again. Returns 2s by default.
	RetryDelay func() time.Duration

	// Max number of retries, 0 means infinite
	Tries int

	// ShouldRetry tells whether error should be retried. Nil defaults to always
	// true.
	ShouldRetry func(error) bool
}

type RetryExhaustedError struct {
	Err error
}

func (err *RetryExhaustedError) Error() string {
	if err == nil || err.Err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("retry count exhausted. Last err: %s", err.Err)
}

// Run will repeatedly retry the proivided fn within the constraints set in the
// retry Config. It will retry until one of the following conditions is met:
// - The provided context is cancelled.
// - The Config.StartTimeout time has passed.
// - The function returns without an error.
// - The maximum number of tries, Config.Tries is exceeded.
// - The function returns with an error that does not satisfy conditions
//   set in the Config.ShouldRetry function.
// If the given function (fn) does not return an error, then Run will return
// nil. Otherwise, Run will return a relevant error.
func (cfg Config) Run(ctx context.Context, fn func(context.Context) error) error {
	retryDelay := func() time.Duration { return 2 * time.Second }
	if cfg.RetryDelay != nil {
		retryDelay = cfg.RetryDelay
	}
	shouldRetry := func(error) bool { return true }
	if cfg.ShouldRetry != nil {
		shouldRetry = cfg.ShouldRetry
	}
	var startTimeout <-chan time.Time // nil chans never unlock !
	if cfg.StartTimeout != 0 {
		startTimeout = time.After(cfg.StartTimeout)
	}

	var err error
	for try := 0; ; try++ {
		if cfg.Tries != 0 && try == cfg.Tries {
			return &RetryExhaustedError{err}
		}
		if err = fn(ctx); err == nil {
			return nil
		}
		if !shouldRetry(err) {
			return err
		}

		log.Print(fmt.Errorf("Retryable error: %s", err))

		select {
		case <-ctx.Done():
			return err
		case <-startTimeout:
			return err
		default:
			time.Sleep(retryDelay())
		}
	}
}

// Backoff is a self contained backoff time calculator. This struct should be
// passed around as a copy as it changes its own fields upon any Backoff call.
// Backoff is not thread safe. For now only a Linear backoff call is
// implemented and the Exponential call will be implemented when needed.
type Backoff struct {
	// Initial time to wait. A Backoff call will change this value.
	InitialBackoff time.Duration
	// Maximum time returned.
	MaxBackoff time.Duration
	// For a Linear backoff, InitialBackoff will be multiplied by Multiplier
	// after each call.
	Multiplier float64
}

// Linear Backoff returns a linearly increasing Duration.
//  n = n * Multiplier.
// the first value of n is InitialBackoff. n is maxed by MaxBackoff.
func (lb *Backoff) Linear() time.Duration {
	wait := lb.InitialBackoff
	lb.InitialBackoff = time.Duration(lb.Multiplier * float64(lb.InitialBackoff))
	if lb.MaxBackoff != 0 && lb.InitialBackoff > lb.MaxBackoff {
		lb.InitialBackoff = lb.MaxBackoff
	}
	return wait
}

// Exponential backoff panics: not implemented, yet.
func (lb *Backoff) Exponential() time.Duration {
	panic("not implemented, yet")
}
