package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestDigitalOceanBoxPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var raw interface{}
	raw = &DigitalOceanBoxPostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("Digitalocean PostProcessor should be a PostProcessor")
	}
}
