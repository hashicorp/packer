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

	// ShouldRetry tells wether error should be retried. Nil defaults to always
	// true.
	ShouldRetry func(error) bool
}

// Run fn until context is cancelled up until StartTimeout time has passed.
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

	for try := 0; ; try++ {
		var err error
		if cfg.Tries != 0 && try == cfg.Tries {
			return err
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

type Backoff struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

func (lb *Backoff) Linear() time.Duration {
	wait := lb.InitialBackoff
	lb.InitialBackoff = time.Duration(lb.Multiplier * float64(lb.InitialBackoff))
	if lb.MaxBackoff != 0 && lb.InitialBackoff > lb.MaxBackoff {
		lb.InitialBackoff = lb.MaxBackoff
	}
	return wait
}
