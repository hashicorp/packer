package common

import (
	"testing"
)

func TestGuestAdditionsConfigPrepare(t *testing.T) {
	c := new(GuestAdditionsConfig)
	var errs []error

	c.GuestAdditionsMode = "disable"
	errs = c.Prepare("none")
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}
