package dockerimport

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}
