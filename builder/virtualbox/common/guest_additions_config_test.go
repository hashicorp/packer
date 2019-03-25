package common

import (
	"testing"
)

func TestGuestAdditionsConfigPrepare(t *testing.T) {
	c := new(GuestAdditionsConfig)
	var errs []error

	c.GuestAdditionsMode = "disable"
	c.Communicator = "none"
	errs = c.Prepare(testConfigTemplate(t))
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}
