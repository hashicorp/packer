package vagrant

import (
	"testing"
)

func TestLXCProvider_impl(t *testing.T) {
	var _ Provider = new(LXCProvider)
}
