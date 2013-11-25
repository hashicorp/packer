package docker

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packer.Builder = new(Builder)
}
