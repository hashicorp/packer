package vagrant

import (
	"testing"
)

func TestDigitalOceanProvider_impl(t *testing.T) {
	var _ Provider = new(DigitalOceanProvider)
}
