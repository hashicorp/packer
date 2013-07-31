package chroot

import "testing"

func TestMountDeviceCleanupFunc_ImplementsCleanupFunc(t *testing.T) {
	var raw interface{}
	raw = new(StepMountDevice)
	if _, ok := raw.(Cleanup); !ok {
		t.Fatalf("cleanup func should be a CleanupFunc")
	}
}
