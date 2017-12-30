package null

import (
	"github.com/hashicorp/packer/packer"
	"testing"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packer.Builder = new(Builder)
}
