package file

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packer.Builder = new(Builder)
}
