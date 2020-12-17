package chroot

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// Cleanup is an interface that some steps implement for early cleanup.
type Cleanup interface {
	CleanupFunc(multistep.StateBag) error
}
