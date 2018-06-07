package common

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
)

func clearEnvVars() {
	os.Unsetenv("AWS_POLL_DELAY_SECONDS")
	os.Unsetenv("AWS_MAX_ATTEMPTS")
	os.Unsetenv("AWS_TIMEOUT_SECONDS")
}

func testGetWaiterOptions(t *testing.T) {
	clearEnvVars()

	// no vars are set
	options := getWaiterOptions()
	if len(options) > 0 {
		t.Fatalf("Did not expect any waiter options to be generated; actual: %#v", options)
	}

	// all vars are set
	os.Setenv("AWS_MAX_ATTEMPTS", "800")
	os.Setenv("AWS_TIMEOUT_SECONDS", "20")
	os.Setenv("AWS_POLL_DELAY_SECONDS", "1")
	options = getWaiterOptions()
	expected := []request.WaiterOption{
		request.WithWaiterDelay(request.ConstantWaiterDelay(time.Duration(1) * time.Second)),
		request.WithWaiterMaxAttempts(800),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}
	clearEnvVars()

	// poll delay is not set
	os.Setenv("AWS_MAX_ATTEMPTS", "800")
	options = getWaiterOptions()
	expected = []request.WaiterOption{
		request.WithWaiterMaxAttempts(800),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}
	clearEnvVars()

	// poll delay is not set but timeout seconds is
	os.Setenv("AWS_TIMEOUT_SECONDS", "20")
	options = getWaiterOptions()
	expected = []request.WaiterOption{
		request.WithWaiterDelay(request.ConstantWaiterDelay(time.Duration(2) * time.Second)),
		request.WithWaiterMaxAttempts(10),
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("expected != actual!! Expected: %#v; Actual: %#v.", expected, options)
	}
	clearEnvVars()
}
