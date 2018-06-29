package common

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
)

func testGetWaiterOptions(t *testing.T) {
	// no vars are set
	envValues := overridableWaitVars{
		envInfo{"AWS_POLL_DELAY_SECONDS", 2, false},
		envInfo{"AWS_MAX_ATTEMPTS", 0, false},
		envInfo{"AWS_TIMEOUT_SECONDS", 300, false},
	}
	options := applyEnvOverrides(envValues)
	if len(options) > 0 {
		t.Fatalf("Did not expect any waiter options to be generated; actual: %#v", options)
	}

	// all vars are set
	envValues = overridableWaitVars{
		envInfo{"AWS_POLL_DELAY_SECONDS", 1, true},
		envInfo{"AWS_MAX_ATTEMPTS", 800, true},
		envInfo{"AWS_TIMEOUT_SECONDS", 20, true},
	}
	options = applyEnvOverrides(envValues)
	expected := []request.WaiterOption{
		request.WithWaiterDelay(request.ConstantWaiterDelay(time.Duration(1) * time.Second)),
		request.WithWaiterMaxAttempts(800),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}

	// poll delay is not set
	envValues = overridableWaitVars{
		envInfo{"AWS_POLL_DELAY_SECONDS", 2, false},
		envInfo{"AWS_MAX_ATTEMPTS", 800, true},
		envInfo{"AWS_TIMEOUT_SECONDS", 300, false},
	}
	options = applyEnvOverrides(envValues)
	expected = []request.WaiterOption{
		request.WithWaiterMaxAttempts(800),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}

	// poll delay is not set but timeout seconds is
	envValues = overridableWaitVars{
		envInfo{"AWS_POLL_DELAY_SECONDS", 2, false},
		envInfo{"AWS_MAX_ATTEMPTS", 0, false},
		envInfo{"AWS_TIMEOUT_SECONDS", 20, true},
	}
	options = applyEnvOverrides(envValues)
	expected = []request.WaiterOption{
		request.WithWaiterDelay(request.ConstantWaiterDelay(time.Duration(2) * time.Second)),
		request.WithWaiterMaxAttempts(10),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}
}
