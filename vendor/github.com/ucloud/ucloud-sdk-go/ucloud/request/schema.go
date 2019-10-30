package request

import (
	"time"
)

// String will return a pointer to string
func String(val string) *string {
	return &val
}

// StringValue will return a string from string pointer
func StringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// Int will return a pointer to int
func Int(val int) *int {
	return &val
}

// IntValue will return a int from int pointer
func IntValue(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// Bool will return a pointer to bool
func Bool(val bool) *bool {
	return &val
}

// BoolValue will return a bool from bool pointer
func BoolValue(ptr *bool) bool {
	if ptr != nil {
		return *ptr
	}
	return false
}

// Float64 will return a pointer to float64
func Float64(val float64) *float64 {
	return &val
}

// Float64Value will return a float64 from float64 pointer
func Float64Value(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0.0
}

// TimeDuration will return a pointer to time.Duration
func TimeDuration(val time.Duration) *time.Duration {
	return &val
}

// TimeDurationValue will return a time.Duration from a time.Duration pointer
func TimeDurationValue(ptr *time.Duration) time.Duration {
	if ptr != nil {
		return *ptr
	}
	return 0
}
