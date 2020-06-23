package retry

import (
	"math"
	"math/rand"
	"time"
)

type BackoffFunc func(attempt int) time.Duration

func BackoffLinearWithJitter(waitBetween time.Duration, jitterFraction float64) BackoffFunc {
	return func(attempt int) time.Duration {
		return jitterAround(waitBetween, jitterFraction)
	}
}

func BackoffExponentialWithJitter(base time.Duration, cap time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		to := getExponentialTimeout(attempt, base)
		// Using float types here, because exponential time can be really big, and converting it to time.Duration may
		// result in undefined behaviour. Its safe conversion, when we have compared it to our 'cap' value.
		if to > float64(cap) {
			to = float64(cap)
		}

		return time.Duration(to/2 + to/2*rand.Float64())
	}
}

func getExponentialTimeout(attempt int, base time.Duration) float64 {
	// pow 3: 50ms, 150ms, 450ms, 1.35s, 4.05s, 12.15s - Now.
	// pow 2: 50ms, 100ms, 200ms, 400ms, 800ms, 1.6s - Previous.
	mult := math.Pow(3, float64(attempt))
	return float64(base) * mult
}

func jitterAround(duration time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (rand.Float64()*2 - 1)
	return time.Duration(float64(duration) * (1 + multiplier))
}
