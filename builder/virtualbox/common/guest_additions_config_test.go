package common

import (
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

func TestGuestAdditionsConfigPrepare(t *testing.T) {
	c := new(GuestAdditionsConfig)
	var errs []error

	c.GuestAdditionsMode = "disable"
	c.Communicator = "none"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}
