package vagrant

import (
	"testing"
)

func TestVMwareProvider_impl(t *testing.T) {
	var _ Provider = new(VMwareProvider)
}
