package chroot

import "testing"

func TestCopyFilesCleanupFunc_ImplementsCleanupFunc(t *testing.T) {
	var raw interface{}
	raw = new(StepCopyFiles)
	if _, ok := raw.(Cleanup); !ok {
		t.Fatalf("cleanup func should be a CleanupFunc")
	}
}
