package main

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
)

func openTTY() (packer.TTY, error) {
	return nil, fmt.Errorf("no TTY available on solaris")
}
