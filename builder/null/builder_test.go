package null

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packer.Builder = new(Builder)
}
