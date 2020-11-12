package chroot

import "testing"

func TestMountExtraCleanupFunc_ImplementsCleanupFunc(t *testing.T) {
	var raw interface{}
	raw = new(StepMountExtra)
	if _, ok := raw.(Cleanup); !ok {
		t.Fatalf("cleanup func should be a CleanupFunc")
	}
}
