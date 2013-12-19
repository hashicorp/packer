package vagrant

import (
	"testing"
)

func TestVBoxProvider_impl(t *testing.T) {
	var _ Provider = new(VBoxProvider)
}
