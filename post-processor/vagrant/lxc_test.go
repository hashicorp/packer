package vagrant

import (
	"testing"
)

func TestLxcProvider_impl(t *testing.T) {
	var _ Provider = new(LxcProvider)
}
