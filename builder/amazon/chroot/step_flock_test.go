package chroot

import (
	"testing"

	"github.com/hashicorp/packer/common/chroot"
)

func TestFlockCleanupFunc_ImplementsCleanupFunc(t *testing.T) {
	var raw interface{}
	raw = new(StepFlock)
	if _, ok := raw.(chroot.Cleanup); !ok {
		t.Fatalf("cleanup func should be a CleanupFunc")
	}
}
