package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestAWSBoxPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var raw interface{}
	raw = &AWSBoxPostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("AWS PostProcessor should be a PostProcessor")
	}
}
