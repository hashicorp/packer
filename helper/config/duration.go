package config

import (
	"fmt"
	"time"
)

// DurationString is a string that represents a time duration.
//
// A DurationString is validated using time.ParseDuration.
//
// An empty string ("") is a valid (0) DurationString. A time.Sleep(0) returns
// immediately.
type DurationString string

// Duration returns the parsed duration.
// Duration panics if d is invalid.
func (d DurationString) Duration() time.Duration {
	if d == "" {
		return 0
	}
	du, err := time.ParseDuration(string(d))
	if err != nil {
		s := fmt.Sprintf("DurationString: Could not parse '%s' : %v", d, err)
		panic(s)
	}
	return du
}

func (d DurationString) Validate() error {
	if d == "" {
		return nil
	}
	_, err := time.ParseDuration(string(d))
	return err
}
