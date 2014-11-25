package vagrant

import (
	"testing"
)

func TestParallelsProvider_impl(t *testing.T) {
	var _ Provider = new(ParallelsProvider)
}
