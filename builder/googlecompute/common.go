package googlecompute

import (
	"fmt"
	"math"
	"time"
)

var RetryExhaustedError error = fmt.Errorf("Function never succeeded in Retry")

// Retry retries a function up to numTries times with exponential backoff.
// If numTries == 0, retry indefinitely. If interval == 0, Retry will not delay retrying and there will be
// no exponential backoff. If maxInterval == 0, maxInterval is set to +Infinity.
// Intervals are in seconds.
// Returns an error if initial > max intervals, if retries are exhausted, or if the passed function returns
// an error.
func Retry(initialInterval float64, maxInterval float64, numTries uint, function func() (bool, error)) error {
	if maxInterval == 0 {
		maxInterval = math.Inf(1)
	} else if initialInterval < 0 || initialInterval > maxInterval {
		return fmt.Errorf("Invalid retry intervals (negative or initial < max). Initial: %f, Max: %f.", initialInterval, maxInterval)
	}

	var err error
	done := false
	interval := initialInterval
	for i := uint(0); !done && (numTries == 0 || i < numTries); i++ {
	done, err = function()
		if err != nil {
			return err
		}
    
		if !done {
			// Retry after delay. Calculate next delay.
			time.Sleep(time.Duration(interval) * time.Second)
			interval = math.Min(interval * 2, maxInterval)
		}
	}

	if !done {
	  return RetryExhaustedError
	}
	return nil
}
